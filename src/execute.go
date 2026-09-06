package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type visibility string

const (
	visPrivate visibility = "private"
	visPublic  visibility = "public"
)

func fillLicense(body, year, owner string) string {
	body = strings.ReplaceAll(body, "[yyyy]", year)
	body = strings.ReplaceAll(body, "[name of copyright owner]", owner)
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), " \t\n") + "\n"
}

func apacheLicenseBody(h *host, year, owner string) string {
	raw, err := h.capture("", "gh", "api", "licenses/apache-2.0")
	body := apache20Fallback
	if err == nil && raw != "" {
		var payload struct {
			Body string `json:"body"`
		}
		if json.Unmarshal([]byte(raw), &payload) == nil && payload.Body != "" {
			body = payload.Body
		}
	}
	return fillLicense(body, year, owner)
}

func askVisibility(h *host) (visibility, error) {
	if !h.stdinTTY || !h.stdoutTTY {
		return "", errors.New("公開範囲の確認が必要なので、端末から --init を実行してください。")
	}
	reader := bufio.NewReader(h.stdin)
	for {
		h.printf("公開範囲はどちらにしますか？ (private / public) ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", errors.New("キャンセルしました。")
		}
		raw := strings.ToLower(strings.TrimSpace(line))
		switch raw {
		case "private", "priv", "1", "非公開":
			return visPrivate, nil
		case "public", "pub", "2", "公開":
			return visPublic, nil
		default:
			h.errorf("private または public を入力してください。\n")
		}
	}
}

func execute(h *host, c *colorizer, p plan) error {
	if p.actionOf(stepGitInit) == actionBlocked {
		return errors.New("不可の項目があるため実行しない。上の計画を解消してから再実行してください。")
	}

	todo := 0
	blocked := 0
	for _, s := range p.steps {
		if s.action == actionDo {
			todo++
		}
		if s.action == actionBlocked {
			blocked++
		}
	}
	if todo == 0 {
		if blocked > 0 {
			return errors.New("不可の項目があるため実行しない。上の計画を解消してから再実行してください。")
		}
		h.printf("すでに整っているので、何もしませんでした。\n")
		return nil
	}

	var vis visibility
	needCreate := false
	for _, s := range p.steps {
		if s.id == stepRemote && s.action == actionDo && p.remoteKind == remoteCreate {
			needCreate = true
			break
		}
	}
	if needCreate {
		if p.fixedVisibility != "" {
			vis = visibility(p.fixedVisibility)
			h.printf("公開範囲: %s（プロファイル）\n", vis)
		} else {
			var err error
			vis, err = askVisibility(h)
			if err != nil {
				return err
			}
			h.printf("公開範囲: %s\n", vis)
		}
	}

	cwd := p.cwd
	created := make([]string, 0, 2)

	for _, s := range p.steps {
		if s.action != actionDo {
			continue
		}
		h.printf("\n%s\n", c.bold("→ "+s.title))
		switch s.id {
		case stepGitInit:
			if err := h.live(cwd, "git", "init", "-b", "main"); err != nil {
				if err2 := h.live(cwd, "git", "init"); err2 != nil {
					return err2
				}
			}
		case stepGitignore:
			if err := os.WriteFile(filepath.Join(cwd, ".gitignore"), []byte(gitignoreTemplate), 0o644); err != nil {
				return err
			}
			created = append(created, ".gitignore")
			h.printf("wrote .gitignore\n")
		case stepLicense:
			year := fmt.Sprintf("%d", h.now().Year())
			owner := p.gitUserName
			if owner == "" {
				owner = p.login
			}
			if owner == "" {
				owner = "Copyright Owner"
			}
			body := apacheLicenseBody(h, year, owner)
			if err := os.WriteFile(filepath.Join(cwd, "LICENSE"), []byte(body), 0o644); err != nil {
				return err
			}
			created = append(created, "LICENSE")
			h.printf("wrote LICENSE (Apache-2.0)\n")
		case stepCommit:
			_, err := h.capture(cwd, "git", "rev-parse", "HEAD")
			hasCommit := err == nil
			if !hasCommit {
				if err := h.live(cwd, "git", "add", "-A"); err != nil {
					return err
				}
				if err := h.live(cwd, "git", "commit", "-m", "Initial commit"); err != nil {
					return err
				}
				break
			}
			for _, file := range created {
				if err := h.live(cwd, "git", "add", file); err != nil {
					return err
				}
			}
			message := "Add bootstrap files"
			if len(created) > 0 {
				message = "Add " + strings.Join(created, " and ")
			}
			if err := h.live(cwd, "git", "commit", "-m", message); err != nil {
				return err
			}
		case stepRemote:
			if p.remoteKind == remoteConnect {
				if p.login == "" {
					return errors.New("login が空")
				}
				raw, err := h.capture(cwd, "gh", "repo", "view", p.login+"/"+p.repoName, "--json", "url,sshUrl")
				if err != nil {
					return err
				}
				var view ghRepoView
				if err := json.Unmarshal([]byte(raw), &view); err != nil {
					return err
				}
				remoteURL := view.SSHURL
				if remoteURL == "" {
					remoteURL = view.URL
				}
				if err := h.live(cwd, "git", "remote", "add", remoteName, remoteURL); err != nil {
					return err
				}
				if err := h.live(cwd, "git", "push", "-u", remoteName, "HEAD"); err != nil {
					return err
				}
				if view.URL != "" {
					h.printf("%s\n", view.URL)
				}
			} else {
				if vis == "" {
					return errors.New("公開範囲が未選択")
				}
				visFlag := "--private"
				if vis == visPublic {
					visFlag = "--public"
				}
				if err := h.live(cwd, "gh", "repo", "create", p.repoName, "--source=.", "--remote="+remoteName, "--push", visFlag); err != nil {
					return err
				}
			}
		}
	}

	h.printf("\n")
	if blocked > 0 {
		return errors.New("一部の処理は不可のため、まだ完了していません。計画を解消してから再実行してください。")
	}
	h.printf("%s\n", c.green("完了"))
	return nil
}
