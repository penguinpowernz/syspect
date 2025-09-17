package main

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/prometheus/procfs"
)

type TCPConnection struct {
	LocalIP    string `json:"local_ip"`
	LocalPort  string `json:"local_port"`
	RemoteIP   string `json:"remote_ip"`
	RemotePort string `json:"remote_port"`
	Protocol   string `json:"protocol"`
	State      string `json:"state"`
}

var stateMap = map[int]string{
	1:  "ESTABLISHED",
	2:  "SYN_SENT",
	3:  "SYN_RECV",
	4:  "FIN_WAIT1",
	5:  "FIN_WAIT2",
	6:  "TIME_WAIT",
	7:  "CLOSE",
	8:  "CLOSE_WAIT",
	9:  "LAST_ACK",
	10: "LISTEN",
	11: "CLOSING",
}

func main() {
	file, err := os.Open("/proc/net/tcp")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // Skip the first line

	proc, err := procfs.NewFS("/proc")
	tcps, err := proc.NetTCP()

	out := []map[string]interface{}{}

	for _, tcp := range tcps {
		if tcp.St == 6 {
			continue
		}
		ttcp := map[string]interface{}{}
		ttcp["remote_ip"] = tcp.RemAddr
		ttcp["remote_port"] = tcp.RemPort
		ttcp["local_ip"] = tcp.LocalAddr
		ttcp["local_port"] = tcp.LocalPort
		ttcp["uid"] = tcp.UID
		ttcp["state"] = stateMap[int(tcp.St)]

		out = append(out, ttcp)
	}

	json.NewEncoder(os.Stdout).Encode(out)
}
