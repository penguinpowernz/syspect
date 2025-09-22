package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEvent(t *testing.T) {
	localIPs = append(localIPs, "127.0.0.1", "192.168.1.10")

	ev := NewEvent("10 6 8 9 127.0.0.1 9273 127.0.0.1 58200")
	assert.Equal(t, ev.Source.String(), "127.0.0.1:9273")
	assert.Equal(t, ev.Dest.String(), "127.0.0.1:58200")
	// assert.True(t, ev.RemoteServer())
	// assert.False(t, ev.RemoteClient())
	assert.Equal(t, ev.protoFam, "tcp6")

	ev = NewEvent("2 6 4 5 23.1.23.33 58200 192.168.1.10 9273")
	assert.Equal(t, ev.Source.String(), "23.1.23.33:58200")
	assert.Equal(t, ev.Dest.String(), "192.168.1.10:9273")
	assert.False(t, ev.RemoteServer())
	assert.True(t, ev.RemoteClient())
	assert.Equal(t, ev.protoFam, "tcp4")

	ev = NewEvent("2 6 2 1 192.168.1.10 58200 23.1.23.33 3000")
	assert.Equal(t, ev.origin.String(), ev.Local.String())
}

func TestAddr(t *testing.T) {
	localIPs = append(localIPs, "127.0.0.1")

	addr := Addr{ip: "127.0.0.1", port: 9273}
	assert.True(t, addr.IsServicePort())
	assert.True(t, addr.IsLocalIP())
	assert.False(t, addr.IsClientPort())
	assert.Equal(t, addr.String(), "127.0.0.1:9273")

	addr = Addr{ip: "127.0.0.1", port: 30000}
	assert.False(t, addr.IsServicePort())
	assert.True(t, addr.IsLocalIP())
	assert.True(t, addr.IsClientPort())
	assert.Equal(t, addr.String(), "127.0.0.1:30000")

	addr = Addr{ip: "237.84.2.178", port: 30000}
	assert.False(t, addr.IsServicePort())
	assert.False(t, addr.IsLocalIP())
	assert.True(t, addr.IsClientPort())
	assert.Equal(t, addr.String(), "237.84.2.178:30000")

	addr = Addr{ip: "237.84.2.178", port: 3000}
	assert.True(t, addr.IsServicePort())
	assert.False(t, addr.IsLocalIP())
	assert.False(t, addr.IsClientPort())
	assert.Equal(t, addr.String(), "237.84.2.178:3000")
}
