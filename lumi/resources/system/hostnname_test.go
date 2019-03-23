package system_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mondoo.io/mondoo/lumi/resources/system"
	"go.mondoo.io/mondoo/motor"
	"go.mondoo.io/mondoo/motor/mock/toml"
	"go.mondoo.io/mondoo/motor/types"
)

func TestHostnameLinux(t *testing.T) {
	trans, err := toml.New(&types.Endpoint{Backend: "mock", Path: "hostname_linux.toml"})
	if err != nil {
		t.Fatal(err)
	}

	m, err := motor.New(trans)
	if err != nil {
		t.Fatal(err)
	}

	hostame, err := system.Hostname(m)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "abefed34cc9c", hostame)
}

func TestHostnameWindows(t *testing.T) {
	trans, err := toml.New(&types.Endpoint{Backend: "mock", Path: "hostname_windows.toml"})
	if err != nil {
		t.Fatal(err)
	}

	m, err := motor.New(trans)
	if err != nil {
		t.Fatal(err)
	}

	hostame, err := system.Hostname(m)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "WIN-ABCDEFGVHLD", hostame)
}
