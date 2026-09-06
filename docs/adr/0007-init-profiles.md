# 0007. --init / --doc / --dev で初期化プロファイルを選ぶ

Date: 2026-09-06
Status: Accepted

## Context

初期化のたびに LICENSE の有無と公開範囲を手で選ぶのは、よくある用途では冗長になる。典型パターンとして次の 2 つを先に決めたい。

- `doc` … LICENSE を置かない。private。
- `dev` … LICENSE を置く。public。

設定はソース側のファイルに書き、ビルド時にバイナリへ埋め込み、配布は従来どおり単一バイナリにしたい。第三者ライブラリは使わない。

CLI は入力しやすさと覚えやすさを優先し、Go の `flag` だけで扱う。実行系フラグ（`--init` / `--doc` / `--dev`）は同時に 2 つ以上付けない。

## Decision

### CLI

```text
grg              # ヘルプ + プレビュー（書き込みなし）
grg --init       # 従来どおり実行（LICENSE は無ければ書く、公開範囲は都度確認）
grg --doc        # doc プロファイルで実行（--init は不要）
grg --dev        # dev プロファイルで実行（--init は不要）
```

- `--doc` / `--dev` はそれ自体が実行フラグ。`--init` は付けない。
- `--init` / `--doc` / `--dev` のうち、指定できるのは 0 個または 1 個。2 つ以上ならエラー。
- 引数パースは標準の `flag` で行う。

### プロファイルの意味

| 呼び出し | LICENSE | 公開範囲 | 公開範囲の確認 |
|----------|---------|----------|----------------|
| （なし） | 書き込みしない（プレビューのみ） | — | — |
| `--init` | 無ければ書く（現行） | 都度確認（現行） | する |
| `--doc` | 書かない（既存があれば触らない） | private | しない |
| `--dev` | 無ければ書く | public | しない |

プロファイル指定時は、その内容をプレビューにも出し、実行時の対話を減らす。`--doc` / `--dev` のときも、実行前に計画を表示してから書き込む（現行の `--init` と同様）。

### 設定ファイル

- 置き場所: `src/profiles/doc.json`, `src/profiles/dev.json`
- 形式: JSON（`encoding/json` のみ）
- スキーマ例:

```json
{
  "license": false,
  "visibility": "private"
}
```

- `license` … `true` なら未作成時に LICENSE を書く。`false` なら LICENSE 手順を skip（削除はしない）
- `visibility` … `"private"` または `"public"`。プロファイル指定時はこれで固定する

フラグ名とファイル名を対応させる（`--doc` → `doc.json`）。設定を変えたら再ビルドが必要。

### 埋め込みと配布

- `//go:embed profiles/*.json` でビルド時に取り込む
- 実行時にカレントやホームの設定ファイルは読まない
- 反映は `make` / `make release` 後のバイナリから

### 計画（プレビュー）への通し方

選んだモードを `inspect` に渡し、LICENSE ステップや公開範囲の説明に反映する。フラグなしのときは現行どおりプレビューのみ。`--init` のときは現行どおり。

## Consequences

- よくある private ドキュメント用と public 開発用を、短いフラグ 1 つで再現できる
- `flag` だけで足りる
- 実行フラグを増やすときは JSON・bool フラグ・「1 つだけ」制約の更新が必要
- 設定変更は JSON 編集 + 再ビルド。実行時カスタムはできない
- Organization 配下や別ライセンス種別はこの ADR の範囲外

## Alternatives considered

- `--init --doc` … `--init` が冗長なので不採用
- `--init doc` … `flag` では扱いづらいので不採用
- `--profile=doc` … 増えやすいが、今回は覚えやすさ優先で `--doc` / `--dev` を採る
- 実行時に `~/.config/...` を読む … 配布物が環境依存になるので不採用
