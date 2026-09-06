package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type cliArgs struct {
	help        bool
	version     bool
	run         bool
	profileName string // "" = --init（対話）、"doc" / "dev" = プロファイル
}

func parseArgs(argv []string) (cliArgs, error) {
	fs := flag.NewFlagSet("grg", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	init := fs.Bool("init", false, "")
	doc := fs.Bool("doc", false, "")
	dev := fs.Bool("dev", false, "")
	showVersion := fs.Bool("version", false, "")
	help := fs.Bool("help", false, "")
	fs.BoolVar(help, "h", false, "")

	if err := fs.Parse(argv); err != nil {
		return cliArgs{}, fmt.Errorf("不明な引数があります")
	}
	if fs.NArg() > 0 {
		return cliArgs{}, fmt.Errorf("不明な引数: %s", fs.Arg(0))
	}

	var out cliArgs
	out.help = *help

	n := 0
	if *init {
		n++
	}
	if *doc {
		n++
	}
	if *dev {
		n++
	}
	if *showVersion {
		n++
	}
	if n > 1 {
		return cliArgs{}, fmt.Errorf("--init / --doc / --dev / --version は同時にどれか 1 つだけ指定してください")
	}

	switch {
	case *showVersion:
		out.version = true
	case *init:
		out.run = true
	case *doc:
		out.run = true
		out.profileName = "doc"
	case *dev:
		out.run = true
		out.profileName = "dev"
	}

	if out.help {
		out.run = false
		out.version = false
		out.profileName = ""
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

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "grg %s\n", versionString())
}

func printUsage(w io.Writer, c *colorizer, cmd string) {
	fmt.Fprintf(w, "%s — カレントディレクトリを GitHub リポジトリとして整える\n", c.bold("get-ready-for-github"))
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", c.bold("使い方"))
	fmt.Fprintf(w, "  %s\n", cmd)
	fmt.Fprintf(w, "      バージョン、ヘルプ、このディレクトリで実際に行われる処理のプレビュー（書き込みなし）\n")
	fmt.Fprintf(w, "  %s --version\n", cmd)
	fmt.Fprintf(w, "      バージョンを表示する\n")
	fmt.Fprintf(w, "  %s --init\n", cmd)
	fmt.Fprintf(w, "      プレビューどおりに実行する。公開範囲（private / public）は都度確認する\n")
	fmt.Fprintf(w, "  %s --doc\n", cmd)
	fmt.Fprintf(w, "      doc プロファイルで実行する（LICENSE なし / private）\n")
	fmt.Fprintf(w, "  %s --dev\n", cmd)
	fmt.Fprintf(w, "      dev プロファイルで実行する（LICENSE あり / public）\n")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", c.bold("前提"))
	fmt.Fprintln(w, "  ・リポジトリ用ディレクトリを作ったうえで、その中で実行する")
	fmt.Fprintln(w, "  ・git / gh が PATH にある")
	fmt.Fprintln(w, "  ・gh は GitHub Enterprise または github.com にログイン済み")
	fmt.Fprintln(w, "  ・リポジトリ名はディレクトリ名。ライセンスは Apache-2.0（--dev / --init）")
	fmt.Fprintln(w, "  ・ホスト・ユーザーは gh の現在の認証先に従う（URL は埋め込まない）")
	fmt.Fprintln(w, "  ・--init / --doc / --dev / --version は同時にどれか 1 つだけ")
	if names := listProfileNames(); len(names) > 0 {
		fmt.Fprintf(w, "  ・プロファイル: %s\n", strings.Join(names, ", "))
	}
}

func run(h *host) int {
	c := &colorizer{enabled: h.stdoutTTY}
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		h.errorf("%s\n", err.Error())
		h.errorf("使い方を見る: %s\n", h.invocation())
		return 2
	}

	if args.version {
		printVersion(h.stdout)
		return 0
	}

	var prof *profileConfig
	if args.profileName != "" {
		cfg, err := loadProfile(args.profileName)
		if err != nil {
			h.errorf("%s\n", err.Error())
			return 2
		}
		prof = &cfg
	}

	if !args.run {
		printVersion(h.stdout)
		printUsage(h.stdout, c, h.invocation())
		h.printf("\n")
	} else {
		h.printf("\n")
	}

	p, err := inspect(h, args.profileName, prof)
	if err != nil {
		h.errorf("%s\n", err.Error())
		return 1
	}
	printPlan(h, c, p, args)
	if !args.run {
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
