package main

import (
	"fmt"
	"io"
	"os"
)

type cliArgs struct {
	init bool
	help bool
}

func parseArgs(argv []string) (cliArgs, error) {
	var out cliArgs
	for _, arg := range argv {
		switch arg {
		case "--init":
			out.init = true
		case "--help", "-h":
			out.help = true
		default:
			return cliArgs{}, fmt.Errorf("不明な引数: %s", arg)
		}
	}
	if out.help {
		out.init = false
	}
	return out, nil
}

type colorizer struct {
	enabled bool
}

func (c *colorizer) wrap(code, s string) string {
	if !c.enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (c *colorizer) bold(s string) string  { return c.wrap("1", s) }
func (c *colorizer) dim(s string) string   { return c.wrap("2", s) }
func (c *colorizer) green(s string) string { return c.wrap("32", s) }
func (c *colorizer) yellow(s string) string {
	return c.wrap("33", s)
}
func (c *colorizer) red(s string) string { return c.wrap("31", s) }

func printUsage(w io.Writer, c *colorizer, cmd string) {
	fmt.Fprintf(w, "%s — カレントディレクトリを GitHub リポジトリとして整える\n", c.bold("get-ready-for-github"))
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", c.bold("使い方"))
	fmt.Fprintf(w, "  %s\n", cmd)
	fmt.Fprintf(w, "      ヘルプと、このディレクトリで実際に行われる処理のプレビュー（書き込みなし）\n")
	fmt.Fprintf(w, "  %s --init\n", cmd)
	fmt.Fprintf(w, "      プレビューどおりに実行する。公開範囲（private / public）は都度確認する\n")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", c.bold("前提"))
	fmt.Fprintln(w, "  ・リポジトリ用ディレクトリを作ったうえで、その中で実行する")
	fmt.Fprintln(w, "  ・git / gh が PATH にある")
	fmt.Fprintln(w, "  ・gh は GitHub Enterprise または github.com にログイン済み")
	fmt.Fprintln(w, "  ・リポジトリ名はディレクトリ名。ライセンスは Apache-2.0")
	fmt.Fprintln(w, "  ・ホスト・ユーザーは gh の現在の認証先に従う（URL は埋め込まない）")
}

func run(h *host) int {
	c := &colorizer{enabled: h.stdoutTTY}
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		h.errorf("%s\n", err.Error())
		h.errorf("使い方を見る: %s\n", h.invocation())
		return 2
	}

	showHelp := !args.init
	if showHelp {
		printUsage(h.stdout, c, h.invocation())
		h.printf("\n")
	} else {
		h.printf("\n")
	}

	p, err := inspect(h)
	if err != nil {
		h.errorf("%s\n", err.Error())
		return 1
	}
	printPlan(h, c, p)
	if !args.init {
		return 0
	}
	h.printf("\n")
	if err := execute(h, c, p); err != nil {
		h.errorf("%s\n", err.Error())
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(defaultHost()))
}
