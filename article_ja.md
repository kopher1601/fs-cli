---
title: "Google Gemini File Search APIをターミナルで管理するTUIツールを作りました"
emoji: "🔍"
type: "tech"
topics: ["go", "gemini", "tui", "rag", "opensource"]
published: true
---

# Google Gemini File Search APIをターミナルで管理するTUIツールを作りました

## 作った経緯

プロジェクトで簡単なRAGを実装するために、Google Geminiの[File Search Tool](https://ai.google.dev/gemini-api/docs/file-search)を選んで使いました。File Search ToolはGoogleがホスティングするRAGシステムで、ファイルをアップロードすれば自動でチャンキング・エンベディング・インデキシングしてくれるため、自分でベクトルDBを構築する必要がありません。

しかし、まだ**ベータ版**のため不便な点が多くありました。

- **ダッシュボードがない** — Storeにどんなファイルがアップロードされているか確認するには、毎回APIを直接呼び出す必要がありました
- **ファイルアップロードが面倒** — 初期テスト時にcurlやスクリプトでファイルを一つずつアップロードする必要がありました
- **削除も不便** — テスト用ファイルを整理する時もdocument nameをいちいち確認してAPIを呼び出す必要がありました

結局、**「ターミナルでk9sのように管理できたらいいのに」** という思いから、自分でTUIツールを作ることにしました。

## fs-cli の紹介

**fs-cli**はGoogle Gemini File Search APIをk9sスタイルのインタラクティブTUIで管理するGo CLIツールです。

**GitHub**: https://github.com/kopher1601/fs-cli

```
 fs-cli │ Stores                                         ?:Help
──────────────────────────────────────────────────────────────────
  NAME          DISPLAY NAME         ACTIVE  PENDING  SIZE     AGE
► abc123def     My Knowledge Base    42      2        1.2 GiB  3 days ago
  xyz789ghi     Test Store           5       0        45 MiB   1 hour ago
──────────────────────────────────────────────────────────────────
 c:create  d:delete  enter:documents  y:detail  /:filter  ?:help
```

### 主な機能

| 機能 | 説明 |
|------|------|
| **Store管理** | 作成、一覧表示、詳細表示、削除 |
| **Document管理** | 一覧表示、詳細表示、削除 |
| **ファイルアップロード** | プログレスバー付きでファイルアップロード |
| **マルチセレクト** | Spaceで複数Documentを選択して一括削除 |
| **Operation追跡** | アップロード等の非同期処理をリアルタイム追跡 |
| **フィルター/ページネーション** | `/`でフィルタリング、`n`/`p`でページ移動 |
| **CJK対応** | 日本語・韓国語等の全角文字幅を正しく計算 |

## アーキテクチャ

### 全体構成

```
fs-cli/
├── cmd/fs-cli/main.go        # エントリーポイント
├── internal/
│   ├── api/                   # Gemini File Search APIクライアント
│   │   ├── client.go          # Base HTTPクライアント（認証、エラーマッピング）
│   │   ├── types.go           # APIリクエスト/レスポンス型
│   │   ├── stores.go          # Store CRUD
│   │   ├── documents.go       # Document取得/削除
│   │   ├── upload.go          # マルチパートファイルアップロード
│   │   └── operations.go      # Operationポーリング
│   ├── config/                # 設定（GEMINI_API_KEY）
│   ├── model/                 # ドメインモデル（API ↔ TUI変換）
│   ├── ui/                    # 共有スタイル、キーバインド、メッセージ型
│   └── tui/
│       ├── app.go             # ルートアプリモデル（ルーティング、ナビゲーション）
│       ├── components/        # 再利用コンポーネント
│       │   ├── table.go       # マルチセレクトテーブル
│       │   ├── breadcrumb.go  # ナビゲーションパス
│       │   ├── flash.go       # 通知メッセージ
│       │   ├── statusbar.go   # キーヒントバー
│       │   └── confirm.go     # 確認ダイアログ
│       └── views/             # 画面別Model
│           ├── stores.go      # Store一覧
│           ├── documents.go   # Document一覧
│           ├── upload.go      # アップロードフォーム
│           ├── operations.go  # Operation追跡
│           └── help.go        # ヘルプ
```

### レイヤー分離

コードを3つのレイヤーに分離しました。

```
┌─────────────┐
│   tui/views  │  ← 画面ロジック（各ビューのModel）
├─────────────┤
│   tui/comp   │  ← 再利用UIコンポーネント（テーブル、フラッシュ等）
├─────────────┤
│     api      │  ← HTTPクライアント（TUI非依存）
└─────────────┘
```

- **apiレイヤー**はTUIと完全に独立しています。`net/http`のみ使用し、Bubble Teaへの依存はありません。
- **uiパッケージ**はスタイル、キーバインド、共有メッセージ型を管理します。循環参照を防ぐため、`tui`と`components`が共通で参照する基盤パッケージとして分離しました。
- **views**はそれぞれ独立した`tea.Model`実装であり、`app.go`がルーターの役割を果たします。

### Bubble Tea MVUパターン

[Bubble Tea](https://github.com/charmbracelet/bubbletea)のModel-View-Updateパターンに従います。

```
キーボード入力 → app.Update() → アクティブビューのUpdate()
                                    │
                                    ├─ ナビゲーションMsg → ビュー切替
                                    ├─ API Cmd → goroutineでHTTPリクエスト
                                    │                    │
                                    │                    v
                                    │               結果Msg返却
                                    │                    │
                                    └─ モデル更新 → View()再レンダリング
```

API呼び出しはすべて`tea.Cmd`でラップされ、非同期で実行されます。

```go
func fetchStores(client *api.Client, pageToken string) tea.Cmd {
    return func() tea.Msg {
        resp, err := client.ListStores(context.Background(), 20, pageToken)
        if err != nil {
            return ErrMsg{Err: err}
        }
        return StoresLoadedMsg{Stores: resp.FileSearchStores}
    }
}
```

### ナビゲーションスタック

k9sのようにビュー間のドリルダウンが可能です。`navStack`に前の状態をpushし、`Esc`/`q`でpopします。

```
Stores → Documents → Document Detail
  ↑          ↑            │
  └──────────┴── Esc ─────┘
```

各エントリにはビュータイプだけでなく、コンテキスト（現在のStore、Document）も一緒に保存されます。

### マルチセレクト一括削除

テーブルコンポーネントに`selected map[int]bool`を追加し、Spaceキーで行をトグルします。削除時に選択されたアイテムがあれば、1件ずつ順次削除しながらリアルタイムで進捗を表示します。

```
ℹ Deleting... 3/10
```

1件ずつ`deleteProgressMsg`を返すチェーン方式で実装し、毎回の削除ごとにUIが更新されます。

### CJK文字幅対応

テーブルのカラム幅計算で`len()`（バイト数）の代わりに`lipgloss.Width()`（表示幅）を使用しています。これにより日本語・韓国語などの全角文字が2カラムとして正確に計算され、レイアウトが崩れません。

## 技術スタック

| 技術 | 用途 |
|------|------|
| [Go](https://go.dev/) | 言語 |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUIフレームワーク（MVUパターン） |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | ターミナルスタイリング |
| [Bubbles](https://github.com/charmbracelet/bubbles) | テキスト入力、スピナー等のコンポーネント |
| [go-humanize](https://github.com/dustin/go-humanize) | バイト/時間のフォーマット |

cobraやviperのような重い依存関係は使用していません。TUI専用ツールのため`flag.Parse()`も不要です。

## 使い方

### インストール

```bash
go install github.com/kopher1601/fs-cli/cmd/fs-cli@latest
```

### 設定

```bash
export GEMINI_API_KEY="your-api-key"
```

APIキーは[Google AI Studio](https://aistudio.google.com/apikey)から取得できます。

### 実行

```bash
fs-cli
```

### 主なキーバインド

| キー | 動作 |
|----|------|
| `↑/k`、`↓/j` | 上下移動 |
| `Enter` | 進入 / 詳細表示 |
| `Esc`/`q` | 戻る / 終了 |
| `c` | Store作成 |
| `d` | 削除（選択時は一括削除） |
| `u` | ファイルアップロード |
| `Space` | マルチセレクト |
| `/` | フィルター |
| `?` | ヘルプ |
| `o` | Operationsビュー |

## Google File Search APIとは？

Google GeminiのFile Search Toolはホスティング型RAGシステムです。

1. **FileSearchStore**を作成（ドキュメントストア）
2. ファイルを**アップロード** → 自動でチャンキング・エンベディング・インデキシング
3. Geminiモデルに`fileSearch`ツールとして接続 → セマンティック検索ベースの回答生成

ストレージとクエリ実行は**無料**で、エンベディング生成コストのみ**$0.15/1Mトークン**です。

### 対応ファイル形式

PDF、Word、Excel、PowerPoint、JSON、XML、YAML、CSV、Markdown、HTML、ソースコード（Python、Java、Go、Rust、C++など）等、100種類以上のファイル形式をサポートしています。

## おわりに

まだベータ版のFile Search APIを使う中で感じた不便さを解消するために作ったツールです。同じ不便さを感じている方々のお役に立てれば幸いです。

Issue、PR、スターすべて歓迎です！

**GitHub**: https://github.com/kopher1601/fs-cli
