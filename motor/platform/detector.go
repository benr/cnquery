package platform

import (
	"errors"

	"github.com/rs/zerolog/log"
	"go.mondoo.io/mondoo/motor/transports"
	"go.mondoo.io/mondoo/motor/transports/arista"
	"go.mondoo.io/mondoo/motor/transports/vsphere"
)

func NewDetector(t transports.Transport) *Detector {
	return &Detector{
		transport: t,
	}
}

type Detector struct {
	transport transports.Transport
	cache     *Platform
}

func (d *Detector) resolveOS() (*Platform, bool) {
	log.Debug().Msg("detector> start resolving the platfrom")
	return operatingSystems.Resolve(d.transport)
}

func (d *Detector) Platform() (*Platform, error) {
	if d.transport == nil {
		return nil, errors.New("cannot detect platform without a transport")
	}

	// check if platform is in cache
	if d.cache != nil {
		return d.cache, nil
	}

	var pi *Platform
	switch pt := d.transport.(type) {
	case *vsphere.Transport:
		identifier, err := pt.Identifier()
		if err != nil {
			return nil, err
		}
		return VspherePlatform(pt, identifier)
	case *arista.Transport:
		return &Platform{
			Name:    "arista-eos",
			Kind:    d.transport.Kind(),
			Runtime: d.transport.Runtime(),
		}, nil
	default:
		var resolved bool
		pi, resolved = d.resolveOS()
		if !resolved {
			return nil, errors.New("could not determine operating system")
		}
		pi.Kind = d.transport.Kind()
		pi.Runtime = d.transport.Runtime()
	}

	// cache value
	d.cache = pi
	return pi, nil
}