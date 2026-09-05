package main

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type action string

const (
	actionDo      action = "do"
	actionSkip    action = "skip"
	actionBlocked action = "blocked"
)

type stepID string

const (
	stepGitInit   stepID = "gitInit"
	stepGitignore stepID = "gitignore"
	stepLicense   stepID = "license"
	stepCommit    stepID = "commit"
	stepRemote    stepID = "remote"
)

type remoteKind string

const (
	remoteNone    remoteKind = "none"
	remoteCreate  remoteKind = "create"
	remoteConnect remoteKind = "connect"
)

const remoteName = "origin"

type step struct {
	id     stepID
	title  string
	action action
	detail string
}

type plan struct {
	cwd          string
	repoName     string
	host         string
	login        string
	gitUserName  string
	gitUserEmail string
	remoteKind   remoteKind
	steps        []step
}

func (p plan) stepByID(id stepID) (step, bool) {
	for _, s := range p.steps {
		if s.id == id {
			return s, true
		}
	}
	return step{}, false
}

func (p plan) actionOf(id stepID) action {
	s, ok := p.stepByID(id)
	if !ok {
		return ""
	}
	return s.action
}

var repoNameRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func isValidRepoName(name string) bool {
	return repoNameRe.MatchString(name) && name != "." && name != ".."
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasGitDir(cwd string) bool {
	return fileExists(filepath.Join(cwd, ".git"))
}

func hasGitignore(cwd string) bool {
	return fileExists(filepath.Join(cwd, ".gitignore"))
}

func hasLicense(cwd string) bool {
	for _, name := range []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "LICENCE"} {
		if fileExists(filepath.Join(cwd, name)) {
			return true
		}
	}
	return false
}

type ghUser struct {
	Login   string `json:"login"`
	HTMLURL string `json:"html_url"`
}

type ghRepoView struct {
	URL    string `json:"url"`
	SSHURL string `json:"sshUrl"`
}

func inspect(h *host) (plan, error) {
	wd, err := h.getwd()
	if err != nil {
		return plan{}, err
	}
	cwd, err := filepath.EvalSymlinks(wd)
	if err != nil {
		cwd = wd
	}
	p := plan{
		cwd:        cwd,
		repoName:   filepath.Base(cwd),
		host:       h.getenv("GH_HOST"),
		remoteKind: remoteNone,
	}

	hasGit := h.hasCommand("git")
	hasGh := h.hasCommand("gh")

	if hasGit {
		p.gitUserName, _ = h.capture(cwd, "git", "config", "user.name")
		p.gitUserEmail, _ = h.capture(cwd, "git", "config", "user.email")
	}

	if hasGh {
		if raw, err := h.capture(cwd, "gh", "api", "user"); err == nil {
			var user ghUser
			if json.Unmarshal([]byte(raw), &user) == nil {
				p.login = user.Login
				if user.HTMLURL != "" {
					if u, err := url.Parse(user.HTMLURL); err == nil && u.Host != "" {
						p.host = u.Host
					}
				}
			}
		}
	}
	if p.host == "" {
		p.host = "github.com"
	}

	insideOtherRepo := false
	if hasGit && !hasGitDir(cwd) {
		if top, err := h.capture(cwd, "git", "rev-parse", "--show-toplevel"); err == nil && top != "" {
			resolved, err := filepath.EvalSymlinks(top)
			if err != nil {
				resolved = filepath.Clean(top)
			}
			if resolved != cwd {
				insideOtherRepo = true
			}
		}
	}

	gitRepo := hasGitDir(cwd)
	switch {
	case !hasGit:
		p.steps = append(p.steps, step{stepGitInit, "git init", actionBlocked, "git が PATH にない"})
	case insideOtherRepo:
		p.steps = append(p.steps, step{stepGitInit, "git init", actionBlocked, "親ディレクトリの Git リポジトリ内なので、誤ってネストしない"})
	case gitRepo:
		p.steps = append(p.steps, step{stepGitInit, "git init", actionSkip, ".git がある"})
	default:
		p.steps = append(p.steps, step{stepGitInit, "git init -b main", actionDo, ".git がないので初期化する"})
	}

	initBlocked := p.actionOf(stepGitInit) == actionBlocked

	switch {
	case initBlocked:
		p.steps = append(p.steps, step{stepGitignore, ".gitignore を作成", actionBlocked, "git init ができないので中止"})
	case hasGitignore(cwd):
		p.steps = append(p.steps, step{stepGitignore, ".gitignore を作成", actionSkip, "すでに存在する"})
	default:
		p.steps = append(p.steps, step{stepGitignore, ".gitignore を作成", actionDo, "src/templates/gitignore の内容を書く"})
	}

	switch {
	case initBlocked:
		p.steps = append(p.steps, step{stepLicense, "LICENSE を作成", actionBlocked, "git init ができないので中止"})
	case hasLicense(cwd):
		p.steps = append(p.steps, step{stepLicense, "LICENSE を作成", actionSkip, "すでに存在する"})
	default:
		p.steps = append(p.steps, step{stepLicense, "LICENSE を作成", actionDo, "Apache-2.0 を gh API から取得して書く（失敗時は内蔵テキスト）"})
	}

	needCommitFiles := p.actionOf(stepGitignore) == actionDo || p.actionOf(stepLicense) == actionDo
	hasCommit := false
	dirty := false
	if gitRepo && hasGit {
		if _, err := h.capture(cwd, "git", "rev-parse", "HEAD"); err == nil {
			hasCommit = true
		}
		if status, err := h.capture(cwd, "git", "status", "--porcelain"); err == nil {
			dirty = status != ""
		}
	}
	willHaveRepo := gitRepo || p.actionOf(stepGitInit) == actionDo
	needInitialCommit := willHaveRepo && !hasCommit

	switch {
	case initBlocked:
		p.steps = append(p.steps, step{stepCommit, "git add / commit", actionBlocked, "git init ができない"})
	case p.gitUserName == "" || p.gitUserEmail == "":
		if needInitialCommit || needCommitFiles {
			p.steps = append(p.steps, step{stepCommit, "git add / commit", actionBlocked, "git config の user.name / user.email が未設定（このツールでは設定しない）"})
		} else {
			p.steps = append(p.steps, step{stepCommit, "git add / commit", actionSkip, "コミット済み"})
		}
	case needInitialCommit:
		p.steps = append(p.steps, step{stepCommit, "git add / commit", actionDo, "初回コミットがまだないので git add -A して Initial commit"})
	case needCommitFiles:
		p.steps = append(p.steps, step{stepCommit, "git add / commit", actionDo, "追加した .gitignore / LICENSE だけをコミットする"})
	case dirty && !hasCommit:
		p.steps = append(p.steps, step{stepCommit, "git add / commit", actionDo, "未コミットの変更がある"})
	default:
		detail := "対象がない"
		if hasCommit {
			detail = "コミット済み"
		}
		p.steps = append(p.steps, step{stepCommit, "git add / commit", actionSkip, detail})
	}

	originOK := false
	originURL := ""
	if gitRepo {
		if out, err := h.capture(cwd, "git", "remote", "get-url", remoteName); err == nil && out != "" {
			originOK = true
			originURL = out
		}
	}
	nameOK := isValidRepoName(p.repoName)
	commitBlocked := p.actionOf(stepCommit) == actionBlocked

	switch {
	case initBlocked:
		p.steps = append(p.steps, step{stepRemote, "gh repo create", actionBlocked, "git init ができないので中止"})
	case originOK:
		p.remoteKind = remoteNone
		p.steps = append(p.steps, step{stepRemote, "gh repo create", actionSkip, remoteName + " がある（" + originURL + "）"})
	case commitBlocked:
		p.steps = append(p.steps, step{stepRemote, "gh repo create", actionBlocked, "コミットできないので push できない"})
	case !hasGh:
		p.steps = append(p.steps, step{stepRemote, "gh repo create", actionBlocked, "gh が PATH にない"})
	case p.login == "":
		p.steps = append(p.steps, step{stepRemote, "gh repo create", actionBlocked, "gh にログインしていない（gh auth status を確認）"})
	case !nameOK:
		p.steps = append(p.steps, step{stepRemote, "gh repo create", actionBlocked, `ディレクトリ名 "` + p.repoName + `" は GitHub のリポジトリ名に使えない`})
	case originOK:
		p.remoteKind = remoteNone
		p.steps = append(p.steps, step{stepRemote, "gh repo create", actionSkip, remoteName + " がある（" + originURL + "）"})
	default:
		raw, err := h.capture(cwd, "gh", "repo", "view", p.login+"/"+p.repoName, "--json", "url,sshUrl")
		if err == nil {
			var view ghRepoView
			if json.Unmarshal([]byte(raw), &view) == nil && view.URL != "" {
				p.remoteKind = remoteConnect
				p.steps = append(p.steps, step{
					id:     stepRemote,
					title:  "git remote add " + remoteName + " && git push",
					action: actionDo,
					detail: "リモートは既にある（" + view.URL + "）。origin を付けて push する",
				})
				break
			}
		}
		p.remoteKind = remoteCreate
		p.steps = append(p.steps, step{
			id:     stepRemote,
			title:  "gh repo create",
			action: actionDo,
			detail: p.host + " の " + p.login + "/" + p.repoName + " を作成して origin を付け、push する（公開範囲は実行時に確認）",
		})
	}

	return p, nil
}

func printPlan(h *host, c *colorizer, p plan) {
	todo, blocked := 0, 0
	for _, s := range p.steps {
		switch s.action {
		case actionDo:
			todo++
		case actionBlocked:
			blocked++
		}
	}

	login := p.login
	if login == "" {
		login = "(未ログイン)"
	}

	h.printf("%s\n", c.bold("このディレクトリでの実行計画"))
	h.printf("  場所:       %s\n", p.cwd)
	h.printf("  リポジトリ: %s\n", p.repoName)
	h.printf("  ホスト:     %s\n", p.host)
	h.printf("  ユーザー:   %s\n", login)
	h.printf("  ライセンス: Apache-2.0\n")
	if p.gitUserName != "" || p.gitUserEmail != "" {
		h.printf("  作者:       %s <%s>\n", p.gitUserName, p.gitUserEmail)
	}
	h.printf("\n")

	for _, s := range p.steps {
		h.printf("  [%s] %s\n", label(c, s.action), s.title)
		h.printf("         %s\n", c.dim(s.detail))
	}
	h.printf("\n")

	switch {
	case blocked > 0:
		h.printf("%s\n", c.red("不可が "+strconv.Itoa(blocked)+" 件ある。--init しても途中で止まる。"))
	case todo == 0:
		h.printf("すでに整っているので、--init しても何もしない。\n")
	default:
		h.printf("実行される処理は %s 件。実際に行うには:\n", strconv.Itoa(todo))
		h.printf("  %s --init\n", h.invocation())
	}
}

func label(c *colorizer, a action) string {
	switch a {
	case actionDo:
		return c.yellow("実行")
	case actionSkip:
		return c.green("済  ")
	default:
		return c.red("不可")
	}
}
