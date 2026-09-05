# 開発者向けメモ

このリポジトリをビルドしたり、生成物の元データを変えたりする人向けです。コマンドの使い方は [README.md](README.md) を見てください。

Go の標準ライブラリだけで実装しています。第三者パッケージは使いません。実行時に呼ぶ外部コマンドは `git` と `gh` だけです。

## 前提

- [Go](https://go.dev/) 1.22 以降（Homebrew なら `brew install go`）
- `make`
- リリースするときだけ [GitHub CLI (`gh`)](https://cli.github.com/)（ログイン済み）

## ビルド

成果物は darwin/arm64 の `grg` です。できたバイナリを Mac に置くときは Go は不要です。

```bash
make
```

PATH の通った場所へ置いて使います。

```bash
install -m 0755 grg "$HOME/bin/grg"
```

ソースから直接走らせる場合:

```bash
go run .
go run . --init
```

## `.gitignore` のカスタム

`grg` が新規プロジェクトに書く `.gitignore` の中身は `templates/gitignore` です。このテキストを編集してから `make`（または `make release`）で作り直すと、以降の生成に反映されます。すでに存在するプロジェクトの `.gitignore` は触りません。

## リリース

GitHub Releases へ載せるときは `make release` だけ実行します。最新の `vX.Y.Z` タグのパッチを 1 つ上げます。タグがまだ無ければ `v0.1.0` です。

```bash
make release
```

桁を上げたいとき:

```bash
make release-minor
make release-major
```

現在の版と、次の patch / minor / major を見る場合:

```bash
make show-version
```

作業ツリーがきれいな状態で、タグ作成・darwin/arm64 のビルド・`gh release create` まで行います。Asset 名は `grg` です。

## 設計メモ

判断の記録は `docs/adr/`、作業ログは `docs/devlog/` にあります。
