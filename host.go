package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type host struct {
	lookPath  func(string) (string, error)
	capture   func(cwd, name string, args ...string) (string, error)
	live      func(cwd, name string, args ...string) error
	getwd     func() (string, error)
	getenv    func(string) string
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	stdinTTY  bool
	stdoutTTY bool
	now       func() time.Time
	args0     string
}

func defaultHost() *host {
	stdoutTTY := isCharDevice(os.Stdout)
	return &host{
		lookPath:  exec.LookPath,
		capture:   captureCmd,
		live:      liveCmd,
		getwd:     os.Getwd,
		getenv:    os.Getenv,
		stdin:     os.Stdin,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		stdinTTY:  isCharDevice(os.Stdin),
		stdoutTTY: stdoutTTY,
		now:       time.Now,
		args0:     os.Args[0],
	}
}

func isCharDevice(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func captureCmd(cwd, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return out, fmt.Errorf("%w: %s", err, msg)
		}
		return out, err
	}
	return out, nil
}

func liveCmd(cwd, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (h *host) hasCommand(name string) bool {
	_, err := h.lookPath(name)
	return err == nil
}

func (h *host) invocation() string {
	return filepath.Base(h.args0)
}

func (h *host) printf(format string, args ...any) {
	fmt.Fprintf(h.stdout, format, args...)
}

func (h *host) errorf(format string, args ...any) {
	fmt.Fprintf(h.stderr, format, args...)
}
