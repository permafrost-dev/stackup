//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

type Target struct {
	OS   string
	Arch string
}

func main() {
	targets := []struct {
		OS   string
		Arch string
	}{
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"windows", "amd64"},
	}

	compile := func(platform, architecture string, wg *sync.WaitGroup) {
		defer wg.Done()

		cmd := exec.Command("make", "package")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("GOOS=%s", platform),
			fmt.Sprintf("GOARCH=%s", architecture),
			fmt.Sprintf("GOMAXPROCS=%d", runtime.NumCPU()),
		)

		err := cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		fmt.Printf("finished building %s-%s\n", platform, architecture)
	}

	var wg sync.WaitGroup

	wg.Add(len(targets))

	for _, t := range targets {
		fmt.Printf("starting build for %s-%s\n", t.OS, t.Arch)
		go compile(t.OS, t.Arch, &wg)
	}

	wg.Wait()
}
