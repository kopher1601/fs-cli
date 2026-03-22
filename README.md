# fs-cli

[Google Gemini File Search API](https://ai.google.dev/gemini-api/docs/file-search)를 터미널에서 관리하는 k9s 스타일의 TUI 도구입니다.

Google Gemini File Search APIをターミナルで管理するk9sスタイルのTUIツールです。

---

## Features / 機能

- vim 스타일 키보드 네비게이션 / vimスタイルのキーボードナビゲーション
- FileSearchStore 관리 (생성, 목록, 삭제, 상세) / FileSearchStoreの管理（作成、一覧、削除、詳細）
- Document 관리 (목록, 업로드, 삭제, 상세) / Documentの管理（一覧、アップロード、削除、詳細）
- 멀티 셀렉트로 일괄 삭제 / マルチセレクトで一括削除
- 업로드 프로그레스 바 / アップロードのプログレスバー
- 실시간 Operation 추적 / リアルタイムOperation追跡
- 필터 및 페이지네이션 / フィルターとページネーション
- CJK 문자 폭 대응 / CJK文字幅対応

## Demo

```
 fs-cli │ Stores                                         ?:Help
──────────────────────────────────────────────────────────────────
  NAME          DISPLAY NAME         ACTIVE  PENDING  SIZE     AGE
► abc123def     My Knowledge Base    42      2        1.2 GiB  3 days ago
  xyz789ghi     Test Store           5       0        45 MiB   1 hour ago
──────────────────────────────────────────────────────────────────
 c:create  d:delete  enter:documents  y:detail  /:filter  ?:help
```

## Installation / インストール

### From source / ソースから

```bash
go install github.com/kopher1601/fs-cli/cmd/fs-cli@latest
```

### From release / リリースから

[Releases](https://github.com/kopher1601/fs-cli/releases) 페이지에서 바이너리를 다운로드하세요.

[Releases](https://github.com/kopher1601/fs-cli/releases)ページからバイナリをダウンロードしてください。

### Build from source / ソースからビルド

```bash
git clone https://github.com/kopher1601/fs-cli.git
cd fs-cli
make build
```

## Configuration / 設定

Gemini API 키를 환경 변수로 설정합니다.

Gemini APIキーを環境変数に設定します。

```bash
export GEMINI_API_KEY="your-api-key"
```

API 키는 [Google AI Studio](https://aistudio.google.com/apikey)에서 발급받을 수 있습니다.

APIキーは[Google AI Studio](https://aistudio.google.com/apikey)から取得できます。

## Usage / 使い方

```bash
fs-cli
```

## Keyboard Shortcuts / キーボードショートカット

### Global / グローバル

| Key | Action |
|-----|--------|
| `↑/k` | 위로 이동 / 上へ移動 |
| `↓/j` | 아래로 이동 / 下へ移動 |
| `g` | 맨 위로 / 先頭へ |
| `G` | 맨 아래로 / 末尾へ |
| `n` | 다음 페이지 / 次のページ |
| `p` | 이전 페이지 / 前のページ |
| `/` | 필터 / フィルター |
| `?` | 도움말 / ヘルプ |
| `o` | Operations 뷰 / Operationsビュー |
| `Ctrl+R` | 새로고침 / リフレッシュ |
| `q/Esc` | 뒤로가기 / 종료 — 戻る / 終了 |
| `Ctrl+C` | 강제 종료 / 強制終了 |

### Stores View / Storesビュー

| Key | Action |
|-----|--------|
| `Enter` | Documents 목록으로 진입 / Documents一覧へ |
| `c` | Store 생성 / Store作成 |
| `d` | Store 삭제 / Store削除 |
| `D` | Store 강제 삭제 / Store強制削除 |
| `y` | Store 상세 보기 / Store詳細表示 |

### Documents View / Documentsビュー

| Key | Action |
|-----|--------|
| `Space` | 선택/해제 (멀티 셀렉트) / 選択/解除（マルチセレクト） |
| `Enter` | Document 상세 / Document詳細 |
| `u` | 파일 업로드 / ファイルアップロード |
| `d` | Document 삭제 (선택 시 일괄 삭제) / Document削除（選択時は一括削除） |
| `D` | Document 강제 삭제 / Document強制削除 |

### Upload View / Uploadビュー

| Key | Action |
|-----|--------|
| `Tab/↓` | 다음 필드 / 次のフィールド |
| `Shift+Tab/↑` | 이전 필드 / 前のフィールド |
| `Enter` | 업로드 시작 / アップロード開始 |
| `Esc` | 취소 / キャンセル |

## Architecture / アーキテクチャ

```
fs-cli/
├── cmd/fs-cli/          # 엔트리포인트 / エントリーポイント
├── internal/
│   ├── api/             # Gemini File Search API 클라이언트 / APIクライアント
│   ├── config/          # 설정 관리 / 設定管理
│   ├── model/           # 도메인 모델 / ドメインモデル
│   ├── ui/              # 공유 스타일, 키바인딩, 타입 / スタイル、キーバインド、型
│   └── tui/
│       ├── app.go       # 루트 앱 모델 / ルートアプリモデル
│       ├── components/  # 테이블, 상태바 등 / テーブル、ステータスバー等
│       └── views/       # 각 화면 구현 / 各画面の実装
├── Makefile
├── LICENSE              # MIT
└── .goreleaser.yaml
```

### Tech Stack

- [Go](https://go.dev/)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Supported File Types / 対応ファイル形式

Google File Search API가 지원하는 모든 파일 형식을 업로드할 수 있습니다.

Google File Search APIがサポートするすべてのファイル形式をアップロードできます。

- PDF, Word (.docx), Excel (.xlsx), PowerPoint (.pptx)
- JSON, XML, YAML, CSV, TSV
- Plain text, HTML, Markdown, RTF
- Python, Java, Go, Rust, C++ 등 소스 코드 / ソースコード
- ZIP archives

## License / ライセンス

[MIT](LICENSE)
