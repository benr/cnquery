package uptime

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.mondoo.com/cnquery/motor/providers/os"
)

var UnixUptimeRegex = regexp.MustCompile(`^.*up[\s]*(?:(\d+)\s(day[s]*|min[s]*),)*(?:\s+([\d:]+),\s)*\s*(?:(\d+)\suser[s]*,\s)*\s*load\s+average[s]*:\s+(\d+[\.,]\d+)[,\s]+(\d+[\.,]\d+)[,\s]+(\d+[\.,]\d+)\s*$`)

type UnixUptimeResult struct {
	Duration           int64
	Users              int
	LoadOneMinute      float64
	LoadFiveMinutes    float64
	LoadFifteenMinutes float64
}

func ParseUnixUptime(uptime string) (*UnixUptimeResult, error) {
	log.Debug().Str("uptime", uptime).Msg("parse")
	m := UnixUptimeRegex.FindStringSubmatch(uptime)

	if len(m) != 8 {
		return nil, fmt.Errorf("could not parse uptime: %s", uptime)
	}

	var duration int64
	var err error

	if len(m[2]) > 0 {
		// caclulate the time x * days / minutes + hours ( m[1]*m[2] + m[3])
		duration, err = strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return nil, err
		}

		switch m[2] {
		case "day":
			fallthrough
		case "days":
			duration = duration * 24 * int64(time.Hour)
		case "min":
			fallthrough
		case "mins":
			duration = duration * int64(time.Minute)
		}
	}

	// add optional hours
	if len(m[3]) > 0 {
		hours := strings.Split(m[3], ":")
		if len(hours) == 2 {
			// log.Debug().Msg("parse hour")
			hh, err := strconv.ParseInt(hours[0], 10, 64)
			if err != nil {
				return nil, err
			}

			// log.Debug().Msg("parse minutes")
			mm, err := strconv.ParseInt(hours[1], 10, 64)
			if err != nil {
				return nil, err
			}

			duration = duration + hh*int64(time.Hour) + mm*int64(time.Minute)
		} else {
			return nil, fmt.Errorf("could not parse uptime hours: %s", uptime)
		}
	}

	// users is optional and is not returned on alpine
	users := 0
	if len(m[4]) > 0 {
		users, err = strconv.Atoi(m[4])
		if err != nil {
			return nil, err
		}
	}

	loadOneMinute, err := strconv.ParseFloat(strings.Replace(m[5], ",", ".", 1), 64)
	if err != nil {
		return nil, err
	}

	loadFiveMinutes, err := strconv.ParseFloat(strings.Replace(m[6], ",", ".", 1), 64)
	if err != nil {
		return nil, err
	}

	loadFifteenMinutes, err := strconv.ParseFloat(strings.Replace(m[7], ",", ".", 1), 64)
	if err != nil {
		return nil, err
	}

	return &UnixUptimeResult{
		Duration:           duration,
		Users:              users,
		LoadOneMinute:      loadOneMinute,
		LoadFiveMinutes:    loadFiveMinutes,
		LoadFifteenMinutes: loadFifteenMinutes,
	}, nil
}

type Unix struct {
	provider os.OperatingSystemProvider
}

func (s *Unix) Name() string {
	return "Unix Uptime"
}

func (s *Unix) Duration() (time.Duration, error) {
	cmd, err := s.provider.RunCommand("uptime")
	if err != nil {
		return 0, err
	}

	ut, err := s.parse(cmd.Stdout)
	if err != nil {
		return 0, err
	}

	return time.Duration(ut.Duration), nil
}

func (s *Unix) parse(r io.Reader) (*UnixUptimeResult, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return ParseUnixUptime(string(content))
}
