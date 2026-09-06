# 0008. ビルド時に版数を埋め込み grg --version で出す

Date: 2026-09-06
Status: Accepted

## Context

配布した `grg` だけでは、どのリリースかをコマンド自身から確認できない。GitHub Releases のタグとバイナリを対応づけ、単一バイナリのまま版数を知りたい。

## Decision

- `main.version` を `go build -ldflags '-X main.version=...'` で埋め込む。ソース上の初期値は `dev`。
- `make build` は、明示の `VERSION` が無ければ `git describe --tags --always --dirty`（失敗時は `dev`）を使う。
- `make release*` は、切るタグ名を `VERSION` として `make build` に渡し、その版でバイナリを作ってから Release に載せる。
- CLI に `--version` を追加する。`--init` / `--doc` / `--dev` / `--version` は同時にどれか 1 つだけ。
- 引数なし（プレビュー）のときは、先頭に版数を出してからヘルプと計画を出す。

## Consequences

- `grg --version` だけで版が分かる。
- リリースバイナリの版と git タグが一致する。
- ローカルの素の `go build` では `dev` のままになり得る（Makefile 経由を推奨）。
