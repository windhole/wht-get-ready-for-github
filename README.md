# get-ready-for-github

カレントディレクトリを GitHub（github.com または GitHub Enterprise）のリポジトリとして整えるコマンドです。

ディレクトリは先に自分で作っておき、その中で実行します。リポジトリ名はディレクトリ名です。ライセンスは Apache-2.0、公開範囲（private / public）は実行時に確認します。

書き込みは `--init` を付けたときだけ行います。引数なしではヘルプと、**今いるディレクトリで実際に何が起きるか**のプレビューだけを出します。

## 前提

次が PATH にあり、使える状態であること。

- [bun](https://bun.sh/)
- git
- [GitHub CLI (`gh`)](https://cli.github.com/)（対象ホストにログイン済み）

ホストやユーザーは `gh` の現在の認証先に従います。URL は埋め込んでいません。Enterprise を使う場合は、先にそのホストで `gh auth login` しておいてください。

git の `user.name` / `user.email` も、コミットが必要なときはあらかじめ設定しておきます。このコマンドは git config を変更しません。

## 使い方

リポジトリ用ディレクトリに移動してから実行します。スクリプトのパスは、このリポジトリを置いた場所に合わせてください。

```bash
cd /path/to/my-new-project

# プレビュー（何も書き込まない）
bun /path/to/wht-get-ready-for-github/get-ready-for-github.ts

# 実行する（公開範囲を確認してから進む）
bun /path/to/wht-get-ready-for-github/get-ready-for-github.ts --init
```

このリポジトリ自身の中なら、次でも同じです。

```bash
bun get-ready-for-github.ts
bun get-ready-for-github.ts --init
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

親ディレクトリがすでに Git リポジトリのときは、誤ってネストしないよう止まります。

## やらないこと

- GitHub Organization 配下への作成
- README の自動生成
- すでに `origin` があるリポジトリの作り直し
