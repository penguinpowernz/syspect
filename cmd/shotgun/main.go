package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: shotgun <port1> [port2] ...")
		return
	}

	killList := []Proc{}
	for {
		for _, port := range os.Args[1:] {
			procs := findPIDsByPort(port)
			killList = append(killList, procs...)
		}

		for _, p := range killList {
			log.Printf("killing %d: %s", p.Pid, p.Path)
			if err := p.Kill(); err != nil {
				log.Printf("failed to kill %d: %v", p.Pid, err)
			}
		}

		time.Sleep(time.Second)
	}
}

type Proc struct {
	Pid  int
	Path string
}

func (p Proc) Kill() error {
	proc, err := os.FindProcess(p.Pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// Find PIDs with a specific TCP port in LISTEN state
func findPIDsByPort(port string) []Proc {
	proto := "TCP"
	// split udp/80 into port and proto arg
	if strings.Contains(port, "/") {
		parts := strings.SplitN(port, "/", 2)
		port = parts[1]
		proto = strings.ToUpper(parts[0])
	}

	cmd := exec.Command("lsof", "-i"+proto+":"+port, "-s"+proto+":LISTEN", "-t")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("failed to execute lsof command: %v: %v", cmd.Args, err)
		return []Proc{}
	}

	pids := []Proc{}
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, Proc{
			Pid:  pid,
			Path: findProcessPath(pid),
		})
	}

	return pids
}

func findProcessPath(pid int) string {
	p, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return ""
	}
	return p
}
