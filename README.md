# grg

カレントディレクトリを GitHub（github.com または GitHub Enterprise）のリポジトリとして整えるコマンドです。

ディレクトリは先に自分で作っておき、その中で実行します。リポジトリ名はディレクトリ名、ライセンスは Apache-2.0 です。公開範囲（private / public）は実行時に確認します。

書き込みは `--init` を付けたときだけ行います。引数なし（または `--help`）ではヘルプと、**今いるディレクトリで実際に何が起きるか**のプレビューだけを出します。

## 前提

PATH にあれば足りるもの:

- git
- [GitHub CLI (`gh`)](https://cli.github.com/)（対象ホストにログイン済み）

ホストやユーザーは `gh` の現在の認証先に従います。URL は埋め込んでいません。Enterprise なら先にそのホストで `gh auth login` してください。

コミットが必要なときは、git の `user.name` / `user.email` も設定しておきます。このコマンドは git config を変更しません。

## 入れ方

[Releases](https://github.com/windhole/wht-get-ready-for-github/releases) から `grg` を落とし、実行権限を付けて PATH へ置きます（Apple Silicon の macOS 向けです）。

```bash
chmod +x grg
install -m 0755 grg "$HOME/bin/grg"
```

ソースからビルドする場合は [DEVELOPING.md](DEVELOPING.md) を見てください。

## 使い方

リポジトリ用ディレクトリに移動してから実行します。

```bash
cd /path/to/my-new-project

# プレビュー（何も書き込まない）
grg

# 実行する（公開範囲を確認してから進む）
grg --init
```

`--init` のとき、まだリモートが無ければ `private` か `public` を聞きます。端末以外（パイプなど）からは実行できません。

## 実行すること

未実施のものだけ進めます。すでに済んでいる手順は飛ばします。

1. `git init`（未初期化なら。ブランチ名は `main`）
2. `.gitignore` がなければ作成
3. `LICENSE` がなければ Apache-2.0 を作成
4. 初回コミットがなければ `git add` と `git commit`
5. `gh repo create` でリモートを作り、`origin` を付けて push

`gh repo create --source` は `--license` と同時に使えないため、ライセンスファイルは先にローカルへ置きます。

親ディレクトリがすでに Git リポジトリのときは、誤ってネストしないよう止まります。`origin` がすでにあればリモート作成はしません。

## やらないこと

- GitHub Organization 配下への作成
- README の自動生成
- すでに `origin` があるリポジトリの作り直し
