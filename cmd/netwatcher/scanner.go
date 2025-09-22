package main

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

type Scanner struct {
	pid     int
	cancel  func()
	mainCtx context.Context
	onEvent func(string)
}

func NewScanner(pid int) *Scanner {
	sc := &Scanner{pid: pid}
	return sc
}

func (scnr *Scanner) OnEvent(cb func(string)) {
	scnr.onEvent = cb
}

func (scnr *Scanner) Reset(pid int) {
	if scnr.cancel != nil {
		scnr.cancel()
	}
	scnr.pid = pid
	scnr.Start(scnr.mainCtx)
}

func (scnr *Scanner) Start(ctx context.Context) {
	scnr.mainCtx = ctx
	ctx, scnr.cancel = context.WithCancel(ctx)

	bfpcode := `tracepoint:sock:inet_sock_set_state /pid == ` + strconv.Itoa(scnr.pid) + `/ { printf(": %d %d %d %d %s %d %s %d\n", args->family, args->protocol, args->oldstate, args->newstate, ntop(args->saddr), args->sport, ntop(args->daddr), args->dport); }`
	cmd := exec.Command("bpftrace", "-e", bfpcode)

	p, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic("Error starting bpftrace: " + err.Error())
	}
	defer cmd.Process.Kill()

	fmt.Println("👍 started BPF trace")

	go func() {
		if err := cmd.Wait(); err != nil {
			if ctx.Err() != nil {
				return
			}
			panic("bpftrace exited " + err.Error())
		}
	}()

	events := make(chan string)
	go scanEvents(ctx, p, events)

	for {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			close(events)
			p.Close()
			return
		case event := <-events:
			scnr.onEvent(event)
		}
	}
}
