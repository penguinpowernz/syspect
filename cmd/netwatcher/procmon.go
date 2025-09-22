package main

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/process"
)

type ProcessMonitor struct {
	pp       *process.Process
	exe      string
	onChange func()
}

func NewProcessMonitor(pid int) (*ProcessMonitor, error) {
	var err error

	pm := &ProcessMonitor{}

	pm.pp, err = process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}

	pm.exe, err = pm.pp.Exe()
	if err != nil {
		return nil, err
	}

	return pm, nil
}

func (p *ProcessMonitor) Start(ctx context.Context) {
	t1 := time.NewTicker(time.Second).C
	fmt.Println("👀 watching process", p.exe)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t1:
			if p.IsRunning() {
				continue
			}
			fmt.Printf("☠️  process died\n")

			p.onChange()
		}
	}
}

func (p *ProcessMonitor) OnChange(cb func(int)) {
	p.onChange = func() {
		fmt.Printf("🔍 watching for new '%s' process to start\n", p.exe)
		p.waitForNewProcess()
		fmt.Printf("👀 new process started on %d\n", p.pp.Pid)
		cb(int(p.pp.Pid))
	}
}

func (pm *ProcessMonitor) waitForNewProcess() {
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
			if exe == pm.exe {
				pm.pp = p
				pm.exe = exe
				return
			}
		}

		time.Sleep(time.Second)
	}
}

func (p *ProcessMonitor) IsRunning() bool {
	if yes, err := p.pp.IsRunning(); err == nil {
		return yes
	}
	return false
}
