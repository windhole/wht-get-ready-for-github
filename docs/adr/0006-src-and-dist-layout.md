# 0006. ソースは src/、ビルド成果物は dist/ に置く

Date: 2026-09-05
Status: Accepted

## Context

リポジトリ直下に Go のソースと成果物が混ざると見通しが悪い。利用者向けの README と開発用ファイルを分けたうえで、ソースとバイナリの置き場所も固定したい。

## Decision

- Go のソースと `go:embed` 用データ（`.gitignore` テンプレート、ライセンスフォールバック）は `src/` 配下に置く。`go.mod` と `Makefile` はリポジトリ直下のまま。
- `make` / `make release` の成果物は `dist/grg` とする。
- `.gitignore` に `dist/` と `data/` を入れる。ローカル作業用の `data/` はリポジトリに含めない。

## Consequences

- トップディレクトリはドキュメント・設定・`go.mod` 中心になる。
- ビルドは `go build -o dist/grg ./src`。テストは `go test ./src/...`。
- ADR-0005 のテンプレートパスは `src/templates/gitignore` に読み替える。
