# 0004. リリースバージョンは git タグからパッチを自動加算する

Date: 2026-09-05
Status: Accepted

## Context

`make release` のバージョンを人が都度決めると、既存タグと重複したり、飛ばしたりしやすい。第三者のリリースツールは使わず、すでに使っている git だけで事故を減らしたい。ADR-0003 の「VERSION は引数で明示する」を改める。

## Decision

- バージョンの正は `vMAJOR.MINOR.PATCH` 形式の git タグとする。ファイルに版数は持たない。
- `make release` は origin のタグを取ったうえで、最新タグのパッチを +1 した版でタグ付け・リリースする。タグが無ければ `v0.1.0`。
- 普段は `make release` だけ実行する。桁を上げるときは `make release-minor` または `make release-major` を使う。
- 現在のタグと次の版番号は `make show-version` で確認する。


## Consequences

- 連続リリースは `v0.1.0` → `v0.1.1` → `v0.1.2` と機械的に進む。
- リモートのタグを見ないままだと衝突し得るので、リリース前に `git fetch --tags` する。
- 破壊的変更のときだけ `make release-major` を意識すればよい。
