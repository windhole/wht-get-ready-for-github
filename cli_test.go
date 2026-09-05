package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		argv    []string
		want    cliArgs
		wantErr bool
	}{
		{nil, cliArgs{}, false},
		{[]string{"--init"}, cliArgs{init: true}, false},
		{[]string{"--help"}, cliArgs{help: true}, false},
		{[]string{"-h", "--init"}, cliArgs{help: true, init: false}, false},
		{[]string{"--nope"}, cliArgs{}, true},
	}
	for _, tc := range cases {
		got, err := parseArgs(tc.argv)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("parseArgs(%v) expected error", tc.argv)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseArgs(%v): %v", tc.argv, err)
		}
		if got != tc.want {
			t.Fatalf("parseArgs(%v)=%+v want %+v", tc.argv, got, tc.want)
		}
	}
}

func TestIsValidRepoName(t *testing.T) {
	t.Parallel()
	ok := []string{"wht-get-ready-for-github", "foo.bar", "a_b", "Repo1"}
	for _, name := range ok {
		if !isValidRepoName(name) {
			t.Fatalf("%q should be valid", name)
		}
	}
	ng := []string{"", ".", "..", "foo bar", "foo/bar", "あ"}
	for _, name := range ng {
		if isValidRepoName(name) {
			t.Fatalf("%q should be invalid", name)
		}
	}
}

func TestFillLicense(t *testing.T) {
	t.Parallel()
	in := "Copyright [yyyy] [name of copyright owner]\n"
	got := fillLicense(in, "2026", "Ada")
	if !strings.Contains(got, "Copyright 2026 Ada") {
		t.Fatalf("got %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Fatal("expected trailing newline")
	}
}

func TestInspectEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	h := fakeHost(t, dir, func(cwd, name string, args ...string) (string, error) {
		key := strings.Join(append([]string{name}, args...), " ")
		switch key {
		case "git config user.name":
			return "Ada", nil
		case "git config user.email":
			return "ada@example.com", nil
		case "gh api user":
			return `{"login":"ada","html_url":"https://github.com/ada"}`, nil
		default:
			return "", errors.New("not found: " + key)
		}
	})
	p, err := inspect(h)
	if err != nil {
		t.Fatal(err)
	}
	if p.repoName != filepath.Base(dir) {
		t.Fatalf("repoName=%s", p.repoName)
	}
	if p.login != "ada" || p.host != "github.com" {
		t.Fatalf("login=%s host=%s", p.login, p.host)
	}
	assertAction(t, p, stepGitInit, actionDo)
	assertAction(t, p, stepGitignore, actionDo)
	assertAction(t, p, stepLicense, actionDo)
	assertAction(t, p, stepCommit, actionDo)
	assertAction(t, p, stepRemote, actionDo)
	if p.remoteKind != remoteCreate {
		t.Fatalf("remoteKind=%s", p.remoteKind)
	}
}

func TestInspectNestedRepo(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	sub := filepath.Join(parent, "nested")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	h := fakeHost(t, sub, func(cwd, name string, args ...string) (string, error) {
		key := strings.Join(append([]string{name}, args...), " ")
		switch key {
		case "git config user.name":
			return "Ada", nil
		case "git config user.email":
			return "ada@example.com", nil
		case "git rev-parse --show-toplevel":
			return parent, nil
		case "gh api user":
			return `{"login":"ada","html_url":"https://github.com/ada"}`, nil
		default:
			return "", errors.New("not found: " + key)
		}
	})
	p, err := inspect(h)
	if err != nil {
		t.Fatal(err)
	}
	assertAction(t, p, stepGitInit, actionBlocked)
	assertAction(t, p, stepGitignore, actionBlocked)
	assertAction(t, p, stepLicense, actionBlocked)
	assertAction(t, p, stepCommit, actionBlocked)
	assertAction(t, p, stepRemote, actionBlocked)
}

func TestInspectSkipsExistingFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := fakeHost(t, dir, func(cwd, name string, args ...string) (string, error) {
		key := strings.Join(append([]string{name}, args...), " ")
		switch key {
		case "git config user.name":
			return "Ada", nil
		case "git config user.email":
			return "ada@example.com", nil
		case "git rev-parse HEAD":
			return "abc123", nil
		case "git status --porcelain":
			return "", nil
		case "git remote get-url origin":
			return "https://github.com/ada/repo.git", nil
		case "gh api user":
			return `{"login":"ada","html_url":"https://github.com/ada"}`, nil
		default:
			return "", errors.New("not found: " + key)
		}
	})
	p, err := inspect(h)
	if err != nil {
		t.Fatal(err)
	}
	assertAction(t, p, stepGitInit, actionSkip)
	assertAction(t, p, stepGitignore, actionSkip)
	assertAction(t, p, stepLicense, actionSkip)
	assertAction(t, p, stepCommit, actionSkip)
	assertAction(t, p, stepRemote, actionSkip)
}

func TestInspectOriginSkipWithoutLogin(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := fakeHost(t, dir, func(cwd, name string, args ...string) (string, error) {
		key := strings.Join(append([]string{name}, args...), " ")
		switch key {
		case "git config user.name":
			return "Ada", nil
		case "git config user.email":
			return "ada@example.com", nil
		case "git rev-parse HEAD":
			return "abc123", nil
		case "git status --porcelain":
			return "", nil
		case "git remote get-url origin":
			return "https://github.com/ada/repo.git", nil
		default:
			return "", errors.New("not found: " + key)
		}
	})
	h.lookPath = func(name string) (string, error) {
		if name == "git" {
			return "/bin/git", nil
		}
		return "", exec.ErrNotFound
	}
	p, err := inspect(h)
	if err != nil {
		t.Fatal(err)
	}
	assertAction(t, p, stepRemote, actionSkip)
}

func TestAskVisibility(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	h := &host{
		stdin:     strings.NewReader("public\n"),
		stdout:    &out,
		stderr:    &out,
		stdinTTY:  true,
		stdoutTTY: true,
	}
	got, err := askVisibility(h)
	if err != nil {
		t.Fatal(err)
	}
	if got != visPublic {
		t.Fatalf("got %s", got)
	}
}

func TestExecuteWritesGitignoreAndLicense(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var live []string
	h := fakeHost(t, dir, func(cwd, name string, args ...string) (string, error) {
		key := strings.Join(append([]string{name}, args...), " ")
		if key == "gh api licenses/apache-2.0" {
			return "", errors.New("offline")
		}
		if key == "git rev-parse HEAD" {
			return "", errors.New("no commits")
		}
		return "", errors.New("unexpected capture: " + key)
	})
	h.live = func(cwd, name string, args ...string) error {
		live = append(live, strings.Join(append([]string{name}, args...), " "))
		return nil
	}
	h.now = func() time.Time { return time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC) }
	p := plan{
		cwd:          dir,
		repoName:     filepath.Base(dir),
		host:         "github.com",
		login:        "ada",
		gitUserName:  "Ada",
		gitUserEmail: "ada@example.com",
		remoteKind:   remoteNone,
		steps: []step{
			{stepGitInit, "git init -b main", actionDo, ""},
			{stepGitignore, ".gitignore を作成", actionDo, ""},
			{stepLicense, "LICENSE を作成", actionDo, ""},
			{stepCommit, "git add / commit", actionDo, ""},
			{stepRemote, "gh repo create", actionSkip, ""},
		},
	}
	c := &colorizer{}
	if err := execute(h, c, p); err != nil {
		t.Fatal(err)
	}
	if !fileExists(filepath.Join(dir, ".gitignore")) {
		t.Fatal("expected .gitignore")
	}
	body, err := os.ReadFile(filepath.Join(dir, "LICENSE"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "Copyright 2026 Ada") {
		t.Fatalf("LICENSE missing copyright: %s", text[len(text)-400:])
	}
	joined := strings.Join(live, " | ")
	if !strings.Contains(joined, "git init -b main") || !strings.Contains(joined, "git commit -m Initial commit") {
		t.Fatalf("live commands: %v", live)
	}
}

func TestUnknownArgExitMessage(t *testing.T) {
	t.Parallel()
	_, err := parseArgs([]string{"--wat"})
	if err == nil || !strings.Contains(err.Error(), "不明な引数") {
		t.Fatalf("err=%v", err)
	}
}

func fakeHost(t *testing.T, dir string, capture func(cwd, name string, args ...string) (string, error)) *host {
	t.Helper()
	return &host{
		lookPath: func(name string) (string, error) {
			if name == "git" || name == "gh" {
				return "/bin/" + name, nil
			}
			return "", exec.ErrNotFound
		},
		capture: capture,
		live: func(cwd, name string, args ...string) error {
			t.Fatalf("unexpected live: %s %v", name, args)
			return nil
		},
		getwd:     func() (string, error) { return dir, nil },
		getenv:    func(string) string { return "" },
		stdin:     strings.NewReader(""),
		stdout:    &bytes.Buffer{},
		stderr:    &bytes.Buffer{},
		now:       func() time.Time { return time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC) },
		args0:     "get-ready-for-github",
		stdinTTY:  false,
		stdoutTTY: false,
	}
}

func assertAction(t *testing.T, p plan, id stepID, want action) {
	t.Helper()
	got := p.actionOf(id)
	if got != want {
		t.Fatalf("step %s action=%s want %s (plan=%+v)", id, got, want, p.steps)
	}
}
