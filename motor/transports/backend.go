package transports

import (
	"errors"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	SCHEME_LOCAL              = "local"
	SCHEME_WINRM              = "winrm"
	SCHEME_SSH                = "ssh"
	SCHEME_DOCKER             = "docker"
	SCHEME_DOCKER_IMAGE       = "docker+image"
	SCHEME_DOCKER_CONTAINER   = "docker+container"
	SCHEME_DOCKER_TAR         = "docker+tar"
	SCHEME_TAR                = "tar"
	SCHEME_K8S                = "k8s"
	SCHEME_GCR                = "gcr" // TODO: this is not part of the transports, merge with cr
	SCHEME_GCP                = "gcp"
	SCHEME_CONTAINER_REGISTRY = "cr"
	SCHEME_AZURE              = "az"
	SCHEME_AWS                = "aws"
	SCHEME_VAGRANT            = "vagrant" // maps to ssh, so no transport required
	SCHEME_MOCK               = "mock"
	SCHEME_VSPHERE            = "vsphere"
	SCHEME_VSPHERE_VM         = "vsphere+vm"
	SCHEME_ARISTA             = "arista"
	SCHEME_MS365              = "ms365"
	SCHEME_IPMI               = "ipmi"
	SCHEME_FS                 = "fs"
	SCHEME_EQUINIX            = "equinix"
)

var TransportBackend_scheme = map[TransportBackend]string{
	TransportBackend_CONNECTION_LOCAL_OS:                SCHEME_LOCAL,
	TransportBackend_CONNECTION_SSH:                     SCHEME_SSH,
	TransportBackend_CONNECTION_WINRM:                   SCHEME_WINRM,
	TransportBackend_CONNECTION_DOCKER:                  SCHEME_DOCKER,
	TransportBackend_CONNECTION_DOCKER_ENGINE_IMAGE:     SCHEME_DOCKER_IMAGE,
	TransportBackend_CONNECTION_DOCKER_ENGINE_CONTAINER: SCHEME_DOCKER_CONTAINER,
	TransportBackend_CONNECTION_AWS_SSM_RUN_COMMAND:     "aws+ssm",
	TransportBackend_CONNECTION_CONTAINER_REGISTRY:      SCHEME_CONTAINER_REGISTRY,
	TransportBackend_CONNECTION_TAR:                     SCHEME_TAR,
	TransportBackend_CONNECTION_MOCK:                    SCHEME_MOCK,
	TransportBackend_CONNECTION_VSPHERE:                 SCHEME_VSPHERE,
	TransportBackend_CONNECTION_ARISTAEOS:               SCHEME_ARISTA,
	TransportBackend_CONNECTION_DOCKER_ENGINE_TAR:       SCHEME_DOCKER_TAR,
	TransportBackend_CONNECTION_AWS:                     SCHEME_AWS,
	TransportBackend_CONNECTION_GCP:                     SCHEME_GCP,
	TransportBackend_CONNECTION_AZURE:                   SCHEME_AZURE,
	TransportBackend_CONNECTION_MS365:                   SCHEME_MS365,
	TransportBackend_CONNECTION_IPMI:                    SCHEME_IPMI,
	TransportBackend_CONNECTION_VSPHERE_VM:              SCHEME_VSPHERE_VM,
	TransportBackend_CONNECTION_FS:                      SCHEME_FS,
	TransportBackend_CONNECTION_K8S:                     SCHEME_K8S,
	TransportBackend_CONNECTION_EQUINIX_METAL:           SCHEME_EQUINIX,
}

var TransportBackend_schemevalue = map[string]TransportBackend{
	SCHEME_LOCAL:              TransportBackend_CONNECTION_LOCAL_OS,
	SCHEME_SSH:                TransportBackend_CONNECTION_SSH,
	SCHEME_WINRM:              TransportBackend_CONNECTION_WINRM,
	SCHEME_DOCKER:             TransportBackend_CONNECTION_DOCKER,
	SCHEME_DOCKER_IMAGE:       TransportBackend_CONNECTION_DOCKER_ENGINE_IMAGE,
	SCHEME_DOCKER_CONTAINER:   TransportBackend_CONNECTION_DOCKER_ENGINE_CONTAINER,
	"aws+ssm":                 TransportBackend_CONNECTION_AWS_SSM_RUN_COMMAND,
	SCHEME_CONTAINER_REGISTRY: TransportBackend_CONNECTION_CONTAINER_REGISTRY,
	SCHEME_TAR:                TransportBackend_CONNECTION_TAR,
	SCHEME_MOCK:               TransportBackend_CONNECTION_MOCK,
	SCHEME_VSPHERE:            TransportBackend_CONNECTION_VSPHERE,
	SCHEME_ARISTA:             TransportBackend_CONNECTION_ARISTAEOS,
	SCHEME_DOCKER_TAR:         TransportBackend_CONNECTION_DOCKER_ENGINE_TAR,
	SCHEME_AWS:                TransportBackend_CONNECTION_AWS,
	SCHEME_GCP:                TransportBackend_CONNECTION_GCP,
	SCHEME_AZURE:              TransportBackend_CONNECTION_AZURE,
	SCHEME_MS365:              TransportBackend_CONNECTION_MS365,
	SCHEME_IPMI:               TransportBackend_CONNECTION_IPMI,
	SCHEME_VSPHERE_VM:         TransportBackend_CONNECTION_VSPHERE_VM,
	SCHEME_FS:                 TransportBackend_CONNECTION_FS,
	SCHEME_K8S:                TransportBackend_CONNECTION_K8S,
	SCHEME_EQUINIX:            TransportBackend_CONNECTION_EQUINIX_METAL,
}

func (x TransportBackend) Scheme() string {
	s, ok := TransportBackend_scheme[x]
	if ok {
		return s
	}
	log.Warn().Str("backend", x.String()).Msg("cannot return scheme for backend")
	return strconv.Itoa(int(x))
}

func MapSchemeBackend(scheme string) (TransportBackend, error) {
	s, ok := TransportBackend_schemevalue[scheme]
	if ok {
		return s, nil
	}

	return TransportBackend_CONNECTION_LOCAL_OS, errors.New("unknown connection scheme: " + scheme)
}
