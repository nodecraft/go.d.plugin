package systemdunits

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/netdata/go.d.plugin/agent/module"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	assert.Implements(t, (*module.Module)(nil), New())
}

func TestSystemdUnits_Init(t *testing.T) {
	tests := map[string]struct {
		config   Config
		wantFail bool
	}{
		"default": {
			config: New().Config,
		},
		"'include' option set": {
			config: Config{
				Include: []string{"*"},
			},
		},
		"'include' option not set": {
			wantFail: true,
			config: Config{
				Include: []string{},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			systemd := New()
			systemd.Config = test.config

			if test.wantFail {
				assert.False(t, systemd.Init())
			} else {
				assert.True(t, systemd.Init())
			}
		})
	}
}

func TestSystemdUnits_Check(t *testing.T) {
	tests := map[string]struct {
		prepare  func() *SystemdUnits
		wantFail bool
	}{
		"successful collection": {
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.Include = []string{"*"}
				systemd.client = prepareOKClient()
				return systemd
			},
		},
		"filter all units": {
			wantFail: true,
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.Include = []string{"*.not_exists"}
				systemd.client = prepareOKClient()
				return systemd
			},
		},
		"error on connect": {
			wantFail: true,
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.client = prepareErrOnConnectClient()
				return systemd
			},
		},
		"error on list units": {
			wantFail: true,
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.client = prepareErrOnListUnitsClient()
				return systemd
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			systemd := test.prepare()
			require.True(t, systemd.Init())

			if test.wantFail {
				assert.False(t, systemd.Check())
			} else {
				assert.True(t, systemd.Check())
			}
		})
	}
}

func TestSystemdUnits_Charts(t *testing.T) {
	assert.NotNil(t, New().Charts())

}

func TestSystemdUnits_Cleanup(t *testing.T) {
	systemd := New()
	systemd.Include = []string{"*"}
	client := prepareOKClient()
	systemd.client = client

	require.True(t, systemd.Init())
	require.NotNil(t, systemd.Collect())
	conn := systemd.conn
	systemd.Cleanup()

	assert.Nil(t, systemd.conn)
	v, _ := conn.(*mockConnection)
	assert.True(t, v.closeCalled)
}

func TestSystemdUnits_Collect(t *testing.T) {
	tests := map[string]struct {
		prepare       func() *SystemdUnits
		wantCollected map[string]int64
	}{
		"collect all unit types": {
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.Include = []string{"*"}
				systemd.client = prepareOKClient()
				return systemd
			},
			wantCollected: map[string]int64{
				"dbus.socket":                              1,
				"dev-disk-by-uuid-DE44-CEE0.device":        1,
				"dev-nvme0n1.device":                       1,
				"docker.socket":                            1,
				"getty-pre.target":                         2,
				"init.scope":                               1,
				"logrotate.timer":                          1,
				"lvm2-lvmetad.socket":                      1,
				"lvm2-lvmpolld.socket":                     1,
				"man-db.timer":                             1,
				"org.cups.cupsd.path":                      1,
				"pamac-cleancache.timer":                   1,
				"pamac-mirrorlist.timer":                   1,
				"proc-sys-fs-binfmt_misc.automount":        1,
				"remote-fs-pre.target":                     2,
				"rpc_pipefs.target":                        2,
				"run-user-1000-gvfs.mount":                 1,
				"run-user-1000.mount":                      1,
				"session-1.scope":                          1,
				"session-2.scope":                          1,
				"session-3.scope":                          1,
				"session-6.scope":                          1,
				"shadow.timer":                             1,
				"sound.target":                             1,
				"sys-devices-virtual-net-loopback1.device": 1,
				"sys-module-fuse.device":                   1,
				"sysinit.target":                           1,
				"system-getty.slice":                       1,
				"system-netctl.slice":                      1,
				"system-systemd-fsck.slice":                1,
				"system.slice":                             1,
				"systemd-ask-password-console.path":        1,
				"systemd-ask-password-wall.path":           1,
				"systemd-ask-password-wall.service":        2,
				"systemd-fsck-root.service":                2,
				"systemd-udevd-kernel.socket":              1,
				"tmp.mount":                                1,
				"user-runtime-dir@1000.service":            1,
				"user.slice":                               1,
				"user@1000.service":                        1,
				"var-lib-nfs-rpc_pipefs.mount":             2,
			},
		},
		"collect only 'service' unit type": {
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.Include = []string{"*.service"}
				systemd.client = prepareOKClient()
				return systemd
			},
			wantCollected: map[string]int64{
				"systemd-ask-password-wall.service": 2,
				"systemd-fsck-root.service":         2,
				"user-runtime-dir@1000.service":     1,
				"user@1000.service":                 1,
			},
		},
		"filter all units": {
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.Include = []string{"*.not_exists"}
				systemd.client = prepareOKClient()
				return systemd
			},
			wantCollected: nil,
		},
		"error on connect": {
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.client = prepareErrOnConnectClient()
				return systemd
			},
			wantCollected: nil,
		},
		"error on list units": {
			prepare: func() *SystemdUnits {
				systemd := New()
				systemd.client = prepareErrOnListUnitsClient()
				return systemd
			},
			wantCollected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			systemd := test.prepare()
			require.True(t, systemd.Init())

			var collected map[string]int64

			for i := 0; i < 10; i++ {
				collected = systemd.Collect()
			}

			assert.Equal(t, test.wantCollected, collected)
			ensureCollectedHasAllChartsDimsVarsIDs(t, systemd, collected)
		})
	}
}

func TestSystemdUnits_connectionReuse(t *testing.T) {
	systemd := New()
	systemd.Include = []string{"*"}
	client := prepareOKClient()
	systemd.client = client
	require.True(t, systemd.Init())

	var collected map[string]int64
	for i := 0; i < 10; i++ {
		collected = systemd.Collect()
	}

	assert.NotEmpty(t, collected)
	assert.Equal(t, 1, client.connectCalls)
}

func ensureCollectedHasAllChartsDimsVarsIDs(t *testing.T, sd *SystemdUnits, collected map[string]int64) {
	for _, chart := range *sd.Charts() {
		if chart.Obsolete {
			continue
		}
		for _, dim := range chart.Dims {
			_, ok := collected[dim.ID]
			assert.Truef(t, ok, "collected metrics has no data for dim '%s' chart '%s'", dim.ID, chart.ID)
		}
		for _, v := range chart.Vars {
			_, ok := collected[v.ID]
			assert.Truef(t, ok, "collected metrics has no data for var '%s' chart '%s'", v.ID, chart.ID)
		}
	}
}

func prepareOKClient() *mockClient {
	return &mockClient{
		conn: &mockConnection{
			units: mockSystemdUnits,
		},
	}
}

func prepareErrOnConnectClient() *mockClient {
	return &mockClient{
		errOnConnect: true,
	}
}

func prepareErrOnListUnitsClient() *mockClient {
	return &mockClient{
		conn: &mockConnection{
			errOnListUnits: true,
		},
	}
}

type mockClient struct {
	conn         systemdConnection
	connectCalls int
	errOnConnect bool
}

func (m *mockClient) connect() (systemdConnection, error) {
	m.connectCalls++
	if m.errOnConnect {
		return nil, errors.New("mock 'connect' error")
	}
	return m.conn, nil
}

type mockConnection struct {
	units          []dbus.UnitStatus
	errOnListUnits bool
	closeCalled    bool
}

func (m *mockConnection) Close() {
	m.closeCalled = true
}

func (m mockConnection) ListUnitsByPatterns(_ []string, patterns []string) ([]dbus.UnitStatus, error) {
	if m.errOnListUnits {
		return nil, errors.New("mock 'ListUnitsByPatterns' error")
	}

	matches := func(name string) bool {
		for _, p := range patterns {
			if ok, _ := filepath.Match(p, name); ok {
				return true
			}
		}
		return false
	}

	var units []dbus.UnitStatus
	for _, unit := range m.units {
		if matches(unit.Name) {
			units = append(units, unit)
		}
	}
	return units, nil
}

var mockSystemdUnits = []dbus.UnitStatus{
	{Name: `proc-sys-fs-binfmt_misc.automount`, LoadState: "loaded", ActiveState: "active"},
	{Name: `dev-nvme0n1.device`, LoadState: "loaded", ActiveState: "active"},
	{Name: `sys-devices-virtual-net-loopback1.device`, LoadState: "loaded", ActiveState: "active"},
	{Name: `sys-module-fuse.device`, LoadState: "loaded", ActiveState: "active"},
	{Name: `dev-disk-by\x2duuid-DE44\x2dCEE0.device`, LoadState: "loaded", ActiveState: "active"},

	{Name: `var-lib-nfs-rpc_pipefs.mount`, LoadState: "loaded", ActiveState: "inactive"},
	{Name: `var.mount`, LoadState: "not-found", ActiveState: "inactive"},
	{Name: `run-user-1000.mount`, LoadState: "loaded", ActiveState: "active"},
	{Name: `tmp.mount`, LoadState: "loaded", ActiveState: "active"},
	{Name: `run-user-1000-gvfs.mount`, LoadState: "loaded", ActiveState: "active"},

	{Name: `org.cups.cupsd.path`, LoadState: "loaded", ActiveState: "active"},
	{Name: `systemd-ask-password-wall.path`, LoadState: "loaded", ActiveState: "active"},
	{Name: `systemd-ask-password-console.path`, LoadState: "loaded", ActiveState: "active"},

	{Name: `init.scope`, LoadState: "loaded", ActiveState: "active"},
	{Name: `session-3.scope`, LoadState: "loaded", ActiveState: "active"},
	{Name: `session-6.scope`, LoadState: "loaded", ActiveState: "active"},
	{Name: `session-1.scope`, LoadState: "loaded", ActiveState: "active"},
	{Name: `session-2.scope`, LoadState: "loaded", ActiveState: "active"},

	{Name: `systemd-fsck-root.service`, LoadState: "loaded", ActiveState: "inactive"},
	{Name: `httpd.service`, LoadState: "not-found", ActiveState: "inactive"},
	{Name: `user-runtime-dir@1000.service`, LoadState: "loaded", ActiveState: "active"},
	{Name: `systemd-ask-password-wall.service`, LoadState: "loaded", ActiveState: "inactive"},
	{Name: `user@1000.service`, LoadState: "loaded", ActiveState: "active"},

	{Name: `user.slice`, LoadState: "loaded", ActiveState: "active"},
	{Name: `system-getty.slice`, LoadState: "loaded", ActiveState: "active"},
	{Name: `system-netctl.slice`, LoadState: "loaded", ActiveState: "active"},
	{Name: `system.slice`, LoadState: "loaded", ActiveState: "active"},
	{Name: `system-systemd\x2dfsck.slice`, LoadState: "loaded", ActiveState: "active"},

	{Name: `lvm2-lvmpolld.socket`, LoadState: "loaded", ActiveState: "active"},
	{Name: `docker.socket`, LoadState: "loaded", ActiveState: "active"},
	{Name: `systemd-udevd-kernel.socket`, LoadState: "loaded", ActiveState: "active"},
	{Name: `dbus.socket`, LoadState: "loaded", ActiveState: "active"},
	{Name: `lvm2-lvmetad.socket`, LoadState: "loaded", ActiveState: "active"},

	{Name: `getty-pre.target`, LoadState: "loaded", ActiveState: "inactive"},
	{Name: `rpc_pipefs.target`, LoadState: "loaded", ActiveState: "inactive"},
	{Name: `remote-fs-pre.target`, LoadState: "loaded", ActiveState: "inactive"},
	{Name: `sysinit.target`, LoadState: "loaded", ActiveState: "active"},
	{Name: `sound.target`, LoadState: "loaded", ActiveState: "active"},

	{Name: `man-db.timer`, LoadState: "loaded", ActiveState: "active"},
	{Name: `pamac-mirrorlist.timer`, LoadState: "loaded", ActiveState: "active"},
	{Name: `pamac-cleancache.timer`, LoadState: "loaded", ActiveState: "active"},
	{Name: `shadow.timer`, LoadState: "loaded", ActiveState: "active"},
	{Name: `logrotate.timer`, LoadState: "loaded", ActiveState: "active"},
}
