# 0002. Go 標準ライブラリのみで CLI を実装する

Date: 2026-09-05
Status: Accepted

## Context

 bun の TypeScript 単一ファイルでも第三者パッケージは使っていなかったが、実行には bun ランタイムが要る。方針を「サードパーティのライブラリに依存しない」「言語標準と official のみ」「macOS では単一バイナリで配りたい」に寄せる。機能は ADR-0001 と同じ（プレビュー、`--init`、git / gh の呼び出し、Apache-2.0 をローカルに置く）で、実行基盤だけを置き換える。

## Decision

- 実装言語は Go。依存モジュールは置かず、標準ライブラリだけを使う（`os/exec`、`encoding/json`、`flag` 相当の自前引数、`os`、`path/filepath`、`embed`）。
- 配布物は `go build` した単一バイナリ。実行時に必要な外部コマンドは従来どおり PATH 上の `git` と `gh` のみ。
- bun / TypeScript のソースと `package.json` は削除する。
- Cobra などの CLI 枠、golang.org/x 配下の拡張パッケージも使わない。
- ホスト URL は埋め込まない。`gh` の認証先に従う。

## Consequences

- 実行側に bun は不要。Mac にはバイナリを 1 つ置けば足りる。
- ビルドには Go ツールチェーンが要る。対象は darwin を想定するが、ソースは Unix 系なら動く。
- git / GitHub API 自体は自前実装せず、既存の `git` / `gh` に任せる。
- 振る舞いを変えない移植なので、プレビューと `--init` の安全策は維持する。
