package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

func scanEvents(ctx context.Context, r io.Reader, events chan string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return
		}

		if ctx.Err() != nil {
			return
		}

		events <- scanner.Text()
	}
}

func parsePid(pid string) int {
	i, err := strconv.Atoi(pid)
	if err != nil {
		panic("pid must be integer: " + err.Error())
	}
	return i
}

func checkWeCanRunBpftrace() bool {
	if _, err := exec.LookPath("bpftrace"); err != nil {
		fmt.Println("bpftrace not installed")
		return false
	}

	if os.Getuid() != 0 {
		fmt.Println("not running as root, can't use bpftrace")
		return false
	}

	return true
}

func contains(list []string, s string) bool {
	for _, l := range list {
		if l == s {
			return true
		}
	}
	return false
}
