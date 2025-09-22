package main

import (
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	_net "github.com/shirou/gopsutil/net"
)

type Event struct {
	protoFam string
	Status   string
	Source   Addr
	Dest     Addr
	incoming bool
	Local    *Addr
	Remote   *Addr
	origin   *Addr
}

func NewEvent(ev string) *Event {
	line := ev
	line = strings.TrimLeft(line, ":")
	line = strings.TrimSpace(line)

	bits := strings.Split(line, " ")
	if len(bits) != 8 {
		return nil
	}

	fam, _ := strconv.Atoi(bits[0])
	proto, _ := strconv.Atoi(bits[1])
	oldState, _ := strconv.Atoi(bits[2])
	newState, _ := strconv.Atoi(bits[3])
	sAddr := bits[4]
	sPort, _ := strconv.Atoi(bits[5])
	dAddr := bits[6]
	dPort, _ := strconv.Atoi(bits[7])

	out := &Event{
		protoFam: parseProtoFamilyRaw(fam, proto),
		Source:   Addr{ip: sAddr, port: sPort},
		Dest:     Addr{ip: dAddr, port: dPort},
		Status:   _net.TCPStatuses["0"+strconv.Itoa(newState)],
	}

	if oldState == 2 || newState == 2 {
		out.origin = &out.Source
	}

	if oldState == 3 || newState == 3 {
		out.origin = &out.Dest
	}

	switch {
	// case isLocalAddr(sAddr) && isLocalAddr(dAddr):
	// out.Local.Service()

	case isLocalAddr(sAddr):
		out.Remote = &out.Source
		out.Local = &out.Dest
		out.incoming = true

	default:
		out.Local = &out.Source
		out.Remote = &out.Dest
		out.incoming = false
	}

	return out
}

func (ev Event) RemoteClient() bool {
	return ev.origin != nil && ev.origin.Equal(ev.Dest) ||
		ev.Remote.Equal(ev.Source) && ev.Dest.IsServicePort() && ev.Source.IsClientPort()
}

func (ev Event) RemoteServer() bool {
	return ev.origin != nil && ev.origin.Equal(ev.Source) ||
		ev.Remote.Equal(ev.Dest) && ev.Source.IsServicePort() && ev.Dest.IsClientPort()
}

type Addr struct {
	ip   string
	port int
}

func (addr Addr) IsServicePort() bool {
	return addr.port < 30000
}

func (addr Addr) IsListeningPort() bool {
	return contains(listenPorts, addr.String())
}

func (addr Addr) IsLocalIP() bool {
	return isLocalAddr(addr.ip)
}

func (addr Addr) IsClientPort() bool {
	return addr.port >= 30000
}

func (addr Addr) Equal(a Addr) bool {
	return addr.ip == a.ip && addr.port == a.port
}

func (addr Addr) String() string {
	return addr.ip + ":" + strconv.Itoa(addr.port)
}

func parseProtoFamilyRaw(fam, proto int) string {
	out := ""
	if proto == syscall.IPPROTO_TCP {
		out += "tcp"
	}

	if proto == syscall.IPPROTO_UDP {
		out += "udp"
	}

	if fam == syscall.AF_INET {
		out += "4"
	}

	if fam == syscall.AF_INET6 {
		out += "6"
	}

	return out
}

func getLocalIps() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

func updateLocalIPs() {
	var err error
	for {
		time.Sleep(time.Minute)
		localIPs, err = getLocalIps()
		if err == nil {
			continue
		}
	}
}

func isLocalAddr(ip string) bool {
	for _, l := range localIPs {
		if l == ip {
			return true
		}
	}
	return false
}
