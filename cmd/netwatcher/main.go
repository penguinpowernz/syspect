package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"

	"github.com/shirou/gopsutil/process"
)

var (
	localIPs    []string
	listenPorts []string
	cache       = map[string]*Event{}
)

func init() {
	localIPs, _ = getLocalIps()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: netwatcher <PID>")
		return
	}

	if !checkWeCanRunBpftrace() {
		return
	}

	pid := parsePid(os.Args[1])

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pm, err := NewProcessMonitor(pid)
	if err != nil {
		panic(err)
	}

	pipe := new(renderPipeline)
	pipe.add(validate)
	pipe.add(parse)
	pipe.add(ignore)
	pipe.add(trackListening)
	pipe.add(resolveDirection)
	pipe.add(caching)
	pipe.add(render)

	// fmt.Println("found local IPs", localIPs)
	go updateLocalIPs()
	scnr := NewScanner(pid)
	scnr.OnEvent(func(ev string) { pipe.run(ev) })
	scanListeningPorts(pid)

	go scnr.Start(ctx)
	pm.OnChange(func(pid int) {
		scanListeningPorts(pid)
		go scnr.Reset(pid)
	})
	pm.Start(ctx)
}

type renderPipeline struct {
	handlers []func(*RenderContext)
}

func (r *renderPipeline) add(handler func(*RenderContext)) {
	r.handlers = append(r.handlers, handler)
}

func (r *renderPipeline) run(evs string) {
	ctx := &RenderContext{evs: evs, ok: true}
	for _, handler := range r.handlers {
		handler(ctx)
	}
}

type RenderContext struct {
	evs string
	ev  *Event
	ok  bool
}

func render(ctx *RenderContext) {
	if !ctx.ok {
		return
	}

	switch {
	case ctx.ev.Status == "LISTEN":
		fmt.Printf("🟢 LISTEN %s %s\n", ctx.ev.protoFam, ctx.ev.Local.String())
		return

	case ctx.ev.incoming:
		fmt.Printf("👈 %s %20s <- %-20s\n", ctx.ev.protoFam, ctx.ev.Local.String(), ctx.ev.Remote.String())
		return

	default:
		fmt.Printf("👉 %s %20s -> %-20s\n", ctx.ev.protoFam, ctx.ev.Local.String(), ctx.ev.Remote.String())
		return
	}
}

func parse(ctx *RenderContext) {
	if !ctx.ok {
		return
	}

	ctx.ev = NewEvent(ctx.evs)
}

func trackListening(ctx *RenderContext) {
	if !ctx.ok {
		return
	}

	if ctx.ev.Status == "LISTEN" {
		listenPorts = append(listenPorts, ctx.ev.Local.String())
	}
}

func caching(ctx *RenderContext) {
	if !ctx.ok {
		return
	}

	x := []string{
		ctx.ev.Local.String(),
		ctx.ev.Remote.String(),
	}
	sort.Strings(x)

	key := fmt.Sprintf("%s-%s", x[0], x[1])
	if _, found := cache[key]; found {
		ctx.ok = false
		return
	}

	cache[key] = ctx.ev
}

func ignore(ctx *RenderContext) {
	if !ctx.ok {
		return
	}

	if ctx.ev.Status == "NONE" {
		ctx.ok = false
	}
}

func validate(ctx *RenderContext) {
	if !ctx.ok {
		return
	}

	if ctx.evs[0] != ':' {
		ctx.ok = false
	}
}

// we have two situations we want to track
// - clients connecting to the listening ports
// - the binary connecting to any non client port
func resolveDirection(ctx *RenderContext) {
	if !ctx.ok {
		return
	}

	if ctx.ev.Status == "LISTEN" {
		return
	}

	src := ctx.ev.Source
	dst := ctx.ev.Dest

	if dst.IsListeningPort() {
		ctx.ev.Local = &ctx.ev.Dest
		ctx.ev.Remote = &ctx.ev.Source
		ctx.ev.incoming = true
		return
	}

	if !src.IsListeningPort() && src.IsLocalIP() {
		ctx.ev.Local = &ctx.ev.Source
		ctx.ev.Remote = &ctx.ev.Dest
		ctx.ev.incoming = false
		return
	}

	ctx.ok = false
}

func scanListeningPorts(pid int) {
	pp, err := process.NewProcess(int32(pid))
	if err != nil {
		panic(err)
	}

	conns, err := pp.Connections()
	if err != nil {
		panic(err)
	}

	listenPorts = listenPorts[:0]
	for _, conn := range conns {
		if conn.Status == "LISTEN" {
			listenPorts = append(listenPorts, conn.Laddr.IP+":"+strconv.Itoa(int(conn.Laddr.Port)))
			if conn.Laddr.IP == "::" {
				listenPorts = append(listenPorts, "127.0.0.1:"+strconv.Itoa(int(conn.Laddr.Port)))
			}
			fmt.Printf("🟢 LISTEN %s\n", conn.Laddr.IP+":"+strconv.Itoa(int(conn.Laddr.Port)))
		}
	}
}
