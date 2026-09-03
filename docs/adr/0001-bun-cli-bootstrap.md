# 0001. bun 単一ファイル CLI で GitHub リポジトリを初期化する

Date: 2026-09-02
Status: Accepted

## Context

GitHub Enterprise（または github.com）に `gh` でログイン済みの前提で、既存ディレクトリからリポジトリ公開までの定型作業（git init / .gitignore / 初回コミット / `gh repo create`）を毎回手でやっている。誤操作を避けつつ、どのマシン・どのホストでも同じ手順で使えるコマンドが欲しい。

`gh repo create --source` は `--license` および `--gitignore` と同時に使えない。ローカルにファイルがある状態でリモートを作るため、ライセンスと gitignore はローカルで用意する必要がある。

## Decision

- ランタイムは bun。依存パッケージは置かず、`#!/usr/bin/env bun` の TypeScript 単一ファイルとし、git / gh は PATH 上のものを呼ぶ。
- 破壊的な処理は `--init` が無いと実行しない。引数なしはヘルプと、カレントディレクトリで実際に起きる手順のプレビュー。
- リポジトリ名はカレントディレクトリ名。可視性（private / public）は `--init` 時に都度確認する。
- LICENSE は Apache-2.0 をローカルに置く。本文は `gh api licenses/apache-2.0` で取得し、失敗時のみ同梱テキストにフォールバックする（ホスト差を吸収するため）。
- リモート作成は `gh repo create <name> --source=. --remote=origin --push --private|--public`。ホスト・ユーザーは `gh` の現在の認証先に従い、URL をハードコードしない。

## Consequences

- bun と gh / git があれば、リポジトリを clone してそのスクリプトを叩くだけで使える。
- GHE で licenses API が無い場合でも LICENSE は作れる。
- `--init` なしでは何も書き込まないので、使い方確認が安全。
- org 配下への作成や README 生成は対象外。必要なら別判断とする。
