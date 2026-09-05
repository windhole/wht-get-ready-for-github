# get-ready-for-github

カレントディレクトリを GitHub（github.com または GitHub Enterprise）のリポジトリとして整えるコマンドです。実行ファイル名は `grg` です。

ディレクトリは先に自分で作っておき、その中で実行します。リポジトリ名はディレクトリ名、ライセンスは Apache-2.0 です。公開範囲（private / public）は実行時に確認します。

書き込みは `--init` を付けたときだけ行います。引数なし（または `--help`）ではヘルプと、**今いるディレクトリで実際に何が起きるか**のプレビューだけを出します。

Go の標準ライブラリだけで実装しています。第三者パッケージは使いません。実行時に呼ぶ外部コマンドは `git` と `gh` だけです。

## 前提

PATH にあれば足りるもの:

- git
- [GitHub CLI (`gh`)](https://cli.github.com/)（対象ホストにログイン済み）

ホストやユーザーは `gh` の現在の認証先に従います。URL は埋め込んでいません。Enterprise なら先にそのホストで `gh auth login` してください。

コミットが必要なときは、git の `user.name` / `user.email` も設定しておきます。このコマンドは git config を変更しません。

## ビルド

ビルドには [Go](https://go.dev/) 1.22 以降と `make` が必要です。成果物は darwin/arm64 の `grg` です。できたバイナリを Mac に置くときは Go は不要です。

```bash
make
```

PATH の通った場所へ置いて使います。

```bash
install -m 0755 grg "$HOME/bin/grg"
```

## 使い方

リポジトリ用ディレクトリに移動してから実行します。

```bash
cd /path/to/my-new-project

# プレビュー（何も書き込まない）
grg

# 実行する（公開範囲を確認してから進む）
grg --init
```

ソースから直接走らせる場合:

```bash
go run .
go run . --init
```

`--init` のとき、まだリモートが無ければ `private` か `public` を聞きます。端末以外（パイプなど）からは実行できません。

## リリース

GitHub Releases へ載せるときは `make release` だけ実行します。最新の `vX.Y.Z` タグのパッチを 1 つ上げます。タグがまだ無ければ `v0.1.0` です。

```bash
make release
```

桁を上げたいときだけ指定します。

```bash
make release BUMP=minor
make release BUMP=major
```

次の版番号だけ見る場合:

```bash
make next-version
```

作業ツリーがきれいな状態で、タグ作成・darwin/arm64 のビルド・`gh release create` まで行います。Asset 名は `grg` です。取る側は落として実行権限を付け、PATH へ置きます。

```bash
chmod +x grg
install -m 0755 grg "$HOME/bin/grg"
```

## 実行すること

未実施のものだけ進めます。すでに済んでいる手順は飛ばします。

1. `git init`（未初期化なら。ブランチ名は `main`）
2. `.gitignore` がなければ作成
3. `LICENSE` がなければ Apache-2.0 を作成
4. 初回コミットがなければ `git add` と `git commit`
5. `gh repo create` でリモートを作り、`origin` を付けて push

`gh repo create --source` は `--license` と同時に使えないため、ライセンスファイルは先にローカルへ置きます。本文は `gh` のライセンス API から取得し、取れなければ内蔵テキストを使います。

親ディレクトリがすでに Git リポジトリのときは、誤ってネストしないよう止まります。`origin` がすでにあればリモート作成はしません。

## やらないこと

- GitHub Organization 配下への作成
- README の自動生成
- すでに `origin` があるリポジトリの作り直し
