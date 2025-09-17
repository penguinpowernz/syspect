package main

// AI SLOP

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: netwatcher <PID>")
		return
	}

	_pid := os.Args[1]

	// Set up a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nExiting...")
		os.Exit(0)
	}()

	pid, err := strconv.Atoi(_pid)
	if err != nil {
		fmt.Printf("Error parsing PID: %v\n", err)
		return
	}

	pp, err := process.NewProcess(int32(pid))
	if err != nil {
		fmt.Printf("Error retrieving process: %v\n", err)
		return
	}

	cache := NewCache()

	t1 := time.NewTicker(time.Second / 10).C
	t2 := time.NewTicker(time.Second / 2).C

	exe, err := pp.Exe()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-sigChan:
			fmt.Println("\nExiting...")
			return

		case <-t1:
			checkConnections(pp, cache)

		case <-t2:
			yes, err := pp.IsRunning()
			if err != nil {
				panic(err)
			}
			if !yes {
				fmt.Println("Process stopped. Waiting for new process...")
				pp = waitForNewProcess(exe)
				fmt.Println("Found new process with PID:", pp.Pid)
				cache = NewCache()
			}
		}
	}
}

func checkConnections(pp *process.Process, cache *Cache) {
	conns, err := pp.Connections()
	if err != nil {
		fmt.Printf("Error retrieving connections: %v\n", err)
		return
	}

	for _, conn := range conns {
		if conn.Status != "ESTABLISHED" {
			continue
		}
		cache.Add(conn)
	}
}

func waitForNewProcess(name string) *process.Process {
	for {
		ps, err := process.Processes()
		if err != nil {
			panic(err)
		}

		for _, p := range ps {
			exe, err := p.Exe()
			if err != nil {
				continue
			}
			if exe == name {
				return p
			}
		}

		time.Sleep(time.Second)
	}
}

type Cache struct {
	entries map[string]net.ConnectionStat
}

func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]net.ConnectionStat),
	}
}

func (c *Cache) Add(conn net.ConnectionStat) {
	key := conn.Raddr.String()
	if _, found := c.entries[key]; found {
		return
	}
	c.entries[key] = conn

	fmt.Printf("Local: %s:%d, Remote: %s:%d, Status: %s\n",
		conn.Laddr.IP, conn.Laddr.Port,
		conn.Raddr.IP, conn.Raddr.Port,
		conn.Status)
}
