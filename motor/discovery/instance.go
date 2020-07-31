package discovery

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.mondoo.io/mondoo/apps/mondoo/cmd/options"
	"go.mondoo.io/mondoo/motor/asset"
	"go.mondoo.io/mondoo/motor/transports"
)

type instanceResolver struct{}

func (k *instanceResolver) Resolve(in *options.VulnOptsAsset, opts *options.VulnOpts) ([]*asset.Asset, error) {
	resolved := []*asset.Asset{}

	refIds := []string{}
	if len(in.ReferenceID) > 0 {
		refIds = []string{in.ReferenceID}
	}

	assetInfo := &asset.Asset{
		// SpaceMrn:     opts.SpaceMrn,
		ReferenceIDs: refIds,
		Labels:       in.Labels,
		State:        asset.State_STATE_ONLINE,
	}

	// parse connection from URI
	// TODO: can we avoid the convertion between asset and motor? should assets use motor connections?
	t := &transports.TransportConfig{}
	err := t.ParseFromURI(in.Connection)
	if err != nil {
		err := errors.Wrapf(err, "cannot connect to %s", in.Connection)
		log.Error().Err(err).Msg("invalid asset connection")
	}

	// copy password from opts asset if it was not encoded in url
	if len(t.Password) == 0 && len(in.Password) > 0 {
		t.Password = in.Password
	}

	t.Sudo = &transports.Sudo{
		Active: opts.Sudo.Active,
	}

	t.IdentityFiles = []string{in.IdentityFile}
	t.Insecure = opts.Insecure
	t.BearerToken = in.BearerToken

	assetInfo.Connections = []*transports.TransportConfig{t}

	if in != nil && len(in.AssetMrn) > 0 {
		assetInfo.Mrn = in.AssetMrn
	}
	resolved = append(resolved, assetInfo)

	return resolved, nil
}