# 0003. make で darwin/arm64 の grg をビルドし GitHub Releases に載せる

Date: 2026-09-05
Status: Accepted

## Context

単一バイナリを GitHub Releases に置きたい。GoReleaser などの第三者ツールは使わず、すでに前提になっている `go` / `git` / `gh` と make だけで済ませたい。配布対象は macOS の Apple Silicon のみで足りる。実行ファイル名は短い `grg` にしたい。

## Decision

- デフォルトの `make`（および `make build`）は `CGO_ENABLED=0 GOOS=darwin GOARCH=arm64` で `grg` を出力する。
- `make release VERSION=vX.Y.Z` は、作業ツリーがきれいなことを確認してからテストとビルドし、そのタグを付けて origin に push し、`gh release create` で `grg` を Asset にする。
- リリース用のバージョンは引数で明示する。自動バンプはしない。
- 対象 OS/Arch は darwin/arm64 のみ。amd64 や linux は作らない。

## Consequences

- Mac では `make` のあと `grg` を PATH に置けば使える。
- リリース作業は `gh` の認証さえあれば追加ツールなしで再現できる。
- Intel Mac 向けバイナリは配らない。必要になったらこの ADR を見直す。
