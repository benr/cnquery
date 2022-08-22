package packages

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	os_provider "go.mondoo.io/mondoo/motor/providers/os"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"go.mondoo.io/mondoo/motor/platform"
)

const (
	RpmPkgFormat = "rpm"
)

var RPM_REGEX = regexp.MustCompile(`^([\w-+]*)\s(\d*|\(none\)):([\w\d-+.:]+)\s([\w\d]*|\(none\))\s(.*)$`)

// ParseRpmPackages parses output from:
// rpm -qa --queryformat '%{NAME} %{EPOCHNUM}:%{VERSION}-%{RELEASE} %{ARCH} %{SUMMARY}\n'
func ParseRpmPackages(input io.Reader) []Package {
	pkgs := []Package{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		m := RPM_REGEX.FindStringSubmatch(line)
		if m != nil {
			var version string
			// only append the epoch if we have a non-zero value
			if m[2] == "0" || strings.TrimSpace(m[2]) == "(none)" {
				version = m[3]
			} else {
				version = m[2] + ":" + m[3]
			}

			arch := m[4]
			// if no arch provided, remove it completely
			if arch == "(none)" {
				arch = ""
			}

			pkgs = append(pkgs, Package{
				Name:        m[1],
				Version:     version,
				Arch:        arch,
				Description: m[5],
				Format:      RpmPkgFormat,
			})
		}
	}
	return pkgs
}

// RpmPkgManager is the pacakge manager for Redhat, CentOS, Oracle and Suse
// it support two modes: runtime where the rpm command is available and static analysis for images (e.g. container tar)
// If the RpmPkgManager is used in static mode, it extracts the rpm database from the system and copies it to the local
// filesystem to run a local rpm command to extract the data. The static analysis is always slower than using the running
// one since more data need to copied. Therefore the runtime check should be preferred over the static analysis
type RpmPkgManager struct {
	provider      os_provider.OperatingSystemProvider
	platform      *platform.Platform
	staticChecked bool
	static        bool
}

func (rpm *RpmPkgManager) Name() string {
	return "Rpm Package Manager"
}

func (rpm *RpmPkgManager) Format() string {
	return RpmPkgFormat
}

// determine if we running against a static image, where we cannot execute the rpm command
// once executed, it caches its result to prevent the execution of the checks many times
func (rpm *RpmPkgManager) isStaticAnalysis() bool {
	if rpm.staticChecked == true {
		return rpm.static
	}

	rpm.static = false

	// check if the rpm command exists, e.g it is not available on tar backend
	c, err := rpm.provider.RunCommand("command -v rpm")
	if err != nil || c.ExitStatus != 0 {
		log.Debug().Msg("mql[packages]> fallback to static rpm package manager")
		rpm.static = true
	}

	// the root problem is that the docker transport (for running containers) cannot easily get the exit code so
	// we cannot always rely on that, a running photon container return non-zero exit code but it will be -1 on the system
	// we probably cannot fix this easily, see dockers approach:
	// https://docs.docker.com/engine/reference/commandline/attach/#get-the-exit-code-of-the-containers-command
	if c != nil {
		rpmCmdPath, err := ioutil.ReadAll(c.Stdout)
		if err != nil || len(rpmCmdPath) == 0 {
			rpm.static = true
		}
	}
	rpm.staticChecked = true
	return rpm.static
}

func (rpm *RpmPkgManager) List() ([]Package, error) {
	if rpm.isStaticAnalysis() {
		return rpm.staticList()
	} else {
		return rpm.runtimeList()
	}
}

func (rpm *RpmPkgManager) Available() (map[string]PackageUpdate, error) {
	if rpm.isStaticAnalysis() {
		return rpm.staticAvailable()
	} else {
		return rpm.runtimeAvailable()
	}
}

func (rpm *RpmPkgManager) queryFormat() string {
	// this format should work everywhere
	// fall-back to epoch instead of epochnum for 6 ish platforms, latest 6 platforms also support epochnum, but we
	// save 1 call by not detecting the available keyword via rpm --querytags
	format := "%{NAME} %{EPOCH}:%{VERSION}-%{RELEASE} %{ARCH} %{SUMMARY}\\n"

	// ATTENTION: EPOCHNUM is only available since later version of rpm in RedHat 6 and Suse 12
	// we can only expect if for rhel 7+, therefore we need to run an extra test
	// be aware that this method is also used for non-redhat systems like suse
	i, err := strconv.ParseInt(rpm.platform.Version, 0, 32)
	if err == nil && (rpm.platform.Name == "centos" || rpm.platform.Name == "redhat") && i >= 7 {
		format = "%{NAME} %{EPOCHNUM}:%{VERSION}-%{RELEASE} %{ARCH} %{SUMMARY}\\n"
	}

	return format
}

func (rpm *RpmPkgManager) runtimeList() ([]Package, error) {
	command := fmt.Sprintf("rpm -qa --queryformat '%s'", rpm.queryFormat())
	cmd, err := rpm.provider.RunCommand(command)
	if err != nil {
		return nil, errors.Wrap(err, "could not read package list")
	}
	return ParseRpmPackages(cmd.Stdout), nil
}

// fetch all available packages, is that working with centos 6?
func (rpm *RpmPkgManager) runtimeAvailable() (map[string]PackageUpdate, error) {
	// python script:
	// import sys;sys.path.insert(0, "/usr/share/yum-cli");import cli;list = cli.YumBaseCli().returnPkgLists(["updates"]);
	// print ''.join(["{\"name\":\""+x.name+"\", \"available\":\""+x.evr+"\",\"arch\":\""+x.arch+"\",\"repo\":\""+x.repo.id+"\"}\n" for x in list.updates]);
	script := "python -c 'import sys;sys.path.insert(0, \"/usr/share/yum-cli\");import cli;list = cli.YumBaseCli().returnPkgLists([\"updates\"]);print \"\".join([ \"{\\\"name\\\":\\\"\"+x.name+\"\\\",\\\"available\\\":\\\"\"+x.evr+\"\\\",\\\"arch\\\":\\\"\"+x.arch+\"\\\",\\\"repo\\\":\\\"\"+x.repo.id+\"\\\"}\\n\" for x in list.updates]);'"

	cmd, err := rpm.provider.RunCommand(script)
	if err != nil {
		log.Debug().Err(err).Msg("mql[packages]> could not read package updates")
		return nil, errors.Wrap(err, "could not read package update list")
	}
	return ParseRpmUpdates(cmd.Stdout)
}

func (rpm *RpmPkgManager) staticList() ([]Package, error) {
	rpmTmpDir, err := ioutil.TempDir(os.TempDir(), "mondoo-rpmdb")
	if err != nil {
		return nil, errors.Wrap(err, "could not create local temp directory")
	}
	log.Debug().Str("path", rpmTmpDir).Msg("cache rpm library locally")
	defer os.RemoveAll(rpmTmpDir)

	fs := rpm.provider.FS()
	afs := &afero.Afero{Fs: fs}

	// on fedora 33+ sqlite is used already, implement new mechanism here
	// if it is stable, we can use it for all rhel
	if rpm.platform != nil && rpm.platform.Name == "fedora" {
		// /var/lib/rpm/rpmdb.sqlite, rpmdb.sqlite-shm and rpmdb.sqlite-wal need to be copied for Fedora 33+
		// We copy the whole /var/lib/rpm directory
		rpmPath := "/var/lib/rpm"
		ok, err := afs.Exists(rpmPath)
		if err != nil {
			return nil, errors.Wrap(err, "rpm directory could not be found")
		}
		if !ok {
			return nil, errors.New("rpm directory could not be found")
		}

		// list directory and copy the content
		wErr := afs.Walk(rpmPath, func(path string, info os.FileInfo, fErr error) error {
			log.Debug().Str("path", path).Str("name", info.Name()).Msg("copy file")
			f, err := fs.Open(path)
			if err != nil {
				return errors.Wrap(err, "could not fetch rpm package list")
			}
			fWriter, err := os.Create(filepath.Join(rpmTmpDir, info.Name()))
			if err != nil {
				log.Error().Err(err).Msg("mql[packages]> could not create tmp file for rpm database")
				return errors.Wrap(err, "could not create local temp file")
			}
			_, err = io.Copy(fWriter, f)
			if err != nil {
				log.Error().Err(err).Msg("mql[packages]> could not copy rpm to tmp file")
				return fmt.Errorf("could not cache rpm package list")
			}
			return nil
		})
		if wErr != nil {
			return nil, errors.Wrap(wErr, "could not fetch rpm package list")
		}
	} else {
		// fetch rpm database file and store it in local tmp file
		// iterate over file paths to check if one exists
		files := []string{
			"/var/lib/rpm/Packages",
			"/usr/lib/sysimage/rpm/Packages", // used on opensuse container
		}
		detectedPath := ""
		for i := range files {
			ok, err := afs.Exists(files[i])
			if err == nil && ok {
				detectedPath = files[i]
				break
			}
		}

		if len(detectedPath) == 0 {
			return nil, errors.Wrap(err, "could not find rpm packages location on : "+rpm.platform.Name)
		}

		f, err := fs.Open(detectedPath)
		if err != nil {
			return nil, errors.Wrap(err, "could not fetch rpm package list")
		}
		fWriter, err := os.Create(filepath.Join(rpmTmpDir, "Packages"))
		if err != nil {
			log.Error().Err(err).Msg("mql[packages]> could not create tmp file for rpm database")
			return nil, errors.Wrap(err, "could not create local temp file")
		}
		_, err = io.Copy(fWriter, f)
		if err != nil {
			log.Error().Err(err).Msg("mql[packages]> could not copy rpm to tmp file")
			return nil, fmt.Errorf("could not cache rpm package list")
		}
	}

	log.Debug().Str("rpmdb", rpmTmpDir).Msg("cached rpm database locally")

	// call local rpm tool to extract the packages
	c := exec.Command("rpm", "--dbpath", rpmTmpDir, "-qa", "--queryformat", rpm.queryFormat())

	stdoutBuffer := bytes.Buffer{}
	stderrBuffer := bytes.Buffer{}

	c.Stdout = &stdoutBuffer
	c.Stderr = &stderrBuffer

	err = c.Run()
	if err != nil {
		log.Error().Err(err).Msg("mql[packages]> could not execute rpm locally")
		return nil, errors.Wrap(err, "could not read package list")
	}

	return ParseRpmPackages(&stdoutBuffer), nil
}

// TODO: Available() not implemented for RpmFileSystemManager
// for now this is not an error since we can easily determine available packages
func (rpm *RpmPkgManager) staticAvailable() (map[string]PackageUpdate, error) {
	return map[string]PackageUpdate{}, nil
}

// SusePkgManager overwrites the normal RPM handler
type SusePkgManager struct {
	RpmPkgManager
}

func (spm *SusePkgManager) Available() (map[string]PackageUpdate, error) {
	cmd, err := spm.provider.RunCommand("zypper --xmlout list-updates")
	if err != nil {
		log.Debug().Err(err).Msg("mql[packages]> could not read package updates")
		return nil, fmt.Errorf("could not read package update list")
	}
	return ParseZypperUpdates(cmd.Stdout)
}