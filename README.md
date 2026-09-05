# get-ready-for-github

カレントディレクトリを GitHub（github.com または GitHub Enterprise）のリポジトリとして整えるコマンドです。Go の標準ライブラリだけで実装しており、第三者パッケージは使いません。

ディレクトリは先に自分で作っておき、その中で実行します。リポジトリ名はディレクトリ名です。ライセンスは Apache-2.0、公開範囲（private / public）は実行時に確認します。

書き込みは `--init` を付けたときだけ行います。引数なしではヘルプと、**今いるディレクトリで実際に何が起きるか**のプレビューだけを出します。

## 前提

次が PATH にあり、使える状態であること。

- git
- [GitHub CLI (`gh`)](https://cli.github.com/)（対象ホストにログイン済み）

ホストやユーザーは `gh` の現在の認証先に従います。URL は埋め込んでいません。Enterprise を使う場合は、先にそのホストで `gh auth login` しておいてください。

git の `user.name` / `user.email` も、コミットが必要なときはあらかじめ設定しておきます。このコマンドは git config を変更しません。

ビルドには [Go](https://go.dev/) 1.22 以降が必要です。実行側の Mac には Go は不要です。

## 使い方

```bash
cd /path/to/my-new-project

# プレビュー（何も書き込まない）
get-ready-for-github

# 実行する（公開範囲を確認してから進む）
get-ready-for-github --init
```

このリポジトリのソースから直接走らせる場合:

```bash
go run .
go run . --init
```

`--init` のとき、まだリモートが無ければ `private` か `public` を聞きます。端末以外（パイプなど）からは実行できません。

## ビルド

macOS（Apple Silicon）向けの単一バイナリ:

```bash
GOOS=darwin GOARCH=arm64 go build -o get-ready-for-github
```

Intel Mac なら `GOARCH=amd64` にします。できたファイルを PATH の通った場所へ置いてください。

## 実行すること

未実施のものだけ進めます。すでに済んでいる手順は飛ばします。

1. `git init`（未初期化なら。ブランチ名は `main`）
2. `.gitignore` がなければ作成
3. `LICENSE` がなければ Apache-2.0 を作成
4. 初回コミットがなければ `git add` と `git commit`
5. `gh repo create` でリモートを作り、`origin` を付けて push

`gh repo create --source` は `--license` と同時に使えないため、ライセンスファイルは先にローカルへ置きます。

親ディレクトリがすでに Git リポジトリのときは、誤ってネストしないよう止まります。

## やらないこと

- GitHub Organization 配下への作成
- README の自動生成
- すでに `origin` があるリポジトリの作り直し
- `git` / `gh` 以外のサードパーティライブラリへの依存
