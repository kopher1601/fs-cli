---
title: "Google Gemini File Search API를 터미널에서 관리하는 TUI 도구를 만들었습니다"
emoji: "🔍"
type: "tech"
topics: ["go", "gemini", "tui", "rag", "opensource"]
published: true
---

# Google Gemini File Search API를 터미널에서 관리하는 TUI 도구를 만들었습니다

## 만들게 된 경위

프로젝트에서 간단한 RAG를 구현하기 위해 Google Gemini의 [File Search Tool](https://ai.google.dev/gemini-api/docs/file-search)을 선택해서 사용했습니다. File Search Tool은 Google이 호스팅하는 RAG 시스템으로, 파일을 업로드하면 자동으로 청킹·임베딩·인덱싱을 해주기 때문에 직접 벡터 DB를 구축할 필요가 없습니다.

하지만 아직 **베타**라서 불편한 점이 많았습니다.

- **대시보드가 없다** — Store에 어떤 파일이 올라가 있는지 확인하려면 매번 API를 직접 호출해야 했습니다
- **파일 업로드가 번거롭다** — 초기 테스트 때 curl이나 스크립트로 파일을 하나씩 올려야 했습니다
- **삭제도 불편하다** — 테스트용 파일을 정리할 때도 document name을 일일이 확인해서 API를 호출해야 했습니다

결국 **"터미널에서 k9s처럼 관리할 수 있으면 좋겠다"** 는 생각으로 직접 TUI 도구를 만들게 되었습니다.

## fs-cli 소개

**fs-cli**는 Google Gemini File Search API를 k9s 스타일의 인터랙티브 TUI로 관리하는 Go CLI 도구입니다.

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

### 주요 기능

| 기능 | 설명 |
|------|------|
| **Store 관리** | 생성, 목록 조회, 상세 보기, 삭제 |
| **Document 관리** | 목록 조회, 상세 보기, 삭제 |
| **파일 업로드** | 프로그레스 바와 함께 파일 업로드 |
| **멀티 셀렉트** | Space로 여러 Document를 선택 후 일괄 삭제 |
| **Operation 추적** | 업로드 등 비동기 작업의 상태를 실시간 추적 |
| **필터/페이지네이션** | `/`로 필터링, `n`/`p`로 페이지 이동 |
| **CJK 대응** | 한국어·일본어 등 전각 문자 폭을 올바르게 계산 |

## 아키텍처

### 전체 구조

```
fs-cli/
├── cmd/fs-cli/main.go        # 엔트리포인트
├── internal/
│   ├── api/                   # Gemini File Search API 클라이언트
│   │   ├── client.go          # Base HTTP 클라이언트 (인증, 에러 매핑)
│   │   ├── types.go           # API 요청/응답 타입
│   │   ├── stores.go          # Store CRUD
│   │   ├── documents.go       # Document 조회/삭제
│   │   ├── upload.go          # 멀티파트 파일 업로드
│   │   └── operations.go      # Operation 폴링
│   ├── config/                # 설정 (GEMINI_API_KEY)
│   ├── model/                 # 도메인 모델 (API ↔ TUI 변환)
│   ├── ui/                    # 공유 스타일, 키바인딩, 메시지 타입
│   └── tui/
│       ├── app.go             # 루트 앱 모델 (라우팅, 네비게이션)
│       ├── components/        # 재사용 컴포넌트
│       │   ├── table.go       # 멀티 셀렉트 테이블
│       │   ├── breadcrumb.go  # 네비게이션 경로
│       │   ├── flash.go       # 알림 메시지
│       │   ├── statusbar.go   # 키 힌트 바
│       │   └── confirm.go     # 확인 다이얼로그
│       └── views/             # 화면별 Model
│           ├── stores.go      # Store 목록
│           ├── documents.go   # Document 목록
│           ├── upload.go      # 업로드 폼
│           ├── operations.go  # Operation 추적
│           └── help.go        # 도움말
```

### 레이어 분리

코드를 3개 레이어로 분리했습니다.

```
┌─────────────┐
│   tui/views  │  ← 화면 로직 (각 뷰의 Model)
├─────────────┤
│   tui/comp   │  ← 재사용 UI 컴포넌트 (테이블, 플래시 등)
├─────────────┤
│     api      │  ← HTTP 클라이언트 (TUI 무관)
└─────────────┘
```

- **api 레이어**는 TUI와 완전히 독립적입니다. `net/http`만 사용하며, Bubble Tea에 대한 의존성이 없습니다.
- **ui 패키지**는 스타일, 키바인딩, 공유 메시지 타입을 관리합니다. 순환 참조를 방지하기 위해 `tui`와 `components`에서 공통으로 참조하는 기반 패키지로 분리했습니다.
- **views**는 각각 독립적인 `tea.Model` 구현체이고, `app.go`가 라우터 역할을 합니다.

### Bubble Tea MVU 패턴

[Bubble Tea](https://github.com/charmbracelet/bubbletea)의 Model-View-Update 패턴을 따릅니다.

```
키보드 입력 → app.Update() → 활성 뷰의 Update()
                                    │
                                    ├─ 네비게이션 Msg → 뷰 전환
                                    ├─ API Cmd → goroutine으로 HTTP 요청
                                    │                    │
                                    │                    v
                                    │               결과 Msg 반환
                                    │                    │
                                    └─ 모델 업데이트 → View() 재렌더링
```

API 호출은 모두 `tea.Cmd`로 래핑되어 비동기로 실행됩니다.

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

### 네비게이션 스택

k9s처럼 뷰 간 드릴다운이 가능합니다. `navStack`에 이전 상태를 push하고, `Esc`/`q`로 pop합니다.

```
Stores → Documents → Document Detail
  ↑          ↑            │
  └──────────┴── Esc ─────┘
```

각 엔트리에는 뷰 타입뿐 아니라 컨텍스트(현재 Store, Document)도 함께 저장됩니다.

### 멀티 셀렉트 일괄 삭제

테이블 컴포넌트에 `selected map[int]bool`을 추가하여 Space 키로 행을 토글합니다. 삭제 시 선택된 항목이 있으면 한 건씩 순차 삭제하면서 실시간 진행도를 표시합니다.

```
ℹ Deleting... 3/10
```

한 건씩 `deleteProgressMsg`를 반환하는 체인 방식으로 구현하여, 매 삭제마다 UI가 갱신됩니다.

### CJK 문자 폭 대응

테이블 칼럼 폭 계산에서 `len()` (바이트 수) 대신 `lipgloss.Width()` (표시 너비)를 사용합니다. 이렇게 하면 한국어·일본어 같은 전각 문자가 2칸으로 정확히 계산되어 레이아웃이 깨지지 않습니다.

## 기술 스택

| 기술 | 용도 |
|------|------|
| [Go](https://go.dev/) | 언어 |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI 프레임워크 (MVU 패턴) |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | 터미널 스타일링 |
| [Bubbles](https://github.com/charmbracelet/bubbles) | 텍스트 입력, 스피너 등 컴포넌트 |
| [go-humanize](https://github.com/dustin/go-humanize) | 바이트/시간 포맷팅 |

cobra나 viper 같은 무거운 의존성은 사용하지 않았습니다. TUI 전용 도구이므로 `flag.Parse()`도 불필요합니다.

## 사용 방법

### 설치

```bash
go install github.com/kopher1601/fs-cli/cmd/fs-cli@latest
```

### 설정

```bash
export GEMINI_API_KEY="your-api-key"
```

API 키는 [Google AI Studio](https://aistudio.google.com/apikey)에서 발급받을 수 있습니다.

### 실행

```bash
fs-cli
```

### 주요 키바인딩

| 키 | 동작 |
|----|------|
| `↑/k`, `↓/j` | 위/아래 이동 |
| `Enter` | 진입 / 상세 보기 |
| `Esc`/`q` | 뒤로가기 / 종료 |
| `c` | Store 생성 |
| `d` | 삭제 (선택 시 일괄 삭제) |
| `u` | 파일 업로드 |
| `Space` | 멀티 셀렉트 |
| `/` | 필터 |
| `?` | 도움말 |
| `o` | Operations 뷰 |

## Google File Search API란?

Google Gemini의 File Search Tool은 호스팅형 RAG 시스템입니다.

1. **FileSearchStore** 생성 (문서 저장소)
2. 파일 **업로드** → 자동으로 청킹·임베딩·인덱싱
3. Gemini 모델에 `fileSearch` 도구로 연결 → 시맨틱 검색 기반 응답 생성

스토리지와 쿼리 실행은 **무료**이고, 임베딩 생성 비용만 **$0.15/1M tokens**입니다.

### 지원 파일 형식

PDF, Word, Excel, PowerPoint, JSON, XML, YAML, CSV, Markdown, HTML, 소스 코드 (Python, Java, Go, Rust, C++ 등) 등 100종 이상의 파일 형식을 지원합니다.

## 마치며

아직 베타인 File Search API를 사용하면서 느꼈던 불편함을 해소하기 위해 만든 도구입니다. 같은 불편함을 겪고 계신 분들에게 도움이 되었으면 합니다.

이슈, PR, 스타 모두 환영합니다!

**GitHub**: https://github.com/kopher1601/fs-cli
