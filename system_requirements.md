너는 Go + HTMX + PostgreSQL 기반의 사내 관리자 웹앱을 설계/구현하는 시니어 풀스택 엔지니어다.

[목표]
- 첨부된 database.sql 파일을 읽고, 그 스키마를 기준으로 전체 구조를 파악한 뒤,
- 회원가입 기능은 필요 없고, 로그인 기능만 구현한다.
- database.sql에 정의된 “모든 테이블”에 대해 관리자 CRUD 페이지를 구현한다.
- 프론트엔드는 HTMX로 구현하고, 서버사이드 렌더링(Go html/template)을 사용한다.
- Go 백엔드는 “초보가 수정하기 쉬운 구조”로 만들어라.
- DB는 PostgreSQL이며, DB 제어는 PostgreSQL MCP를 사용한다.

[DB 접속 정보]
- host: 192.168.0.35
- database: skycontainers
- user: containers_usr
- password: 12345

[필수 기능]
1) 인증/로그인
- 회원가입 없음
- login_id + password 로그인 (email 로그인 아님)
- 로그인 성공 시 세션 기반 인증(쿠키)으로 유지
- 로그아웃 구현
- 라우트 가드(미로그인 시 로그인 화면으로 리다이렉트)
- users 테이블의 role/status를 반영해 접근 제어(최소한: INTERNAL_SUPER_ADMIN / INTERNAL_USER / SUPPLIER_ADMIN / SUPPLIER_USER / READ_ONLY 같은 형태로 확장 가능하게)
- 비밀번호는 bcrypt로 해시하여 저장/검증

2) 관리자 CRUD
- database.sql에 존재하는 모든 테이블에 대해 다음 화면 제공:
  - 목록(list) + 검색(가능하면) + 정렬(가능하면 최소 1개 컬럼)
  - 상세(view)
  - 생성(create)
  - 수정(edit)
  - 삭제(delete) — 실제 삭제가 위험하면 is_active 같은 컬럼이 있는 테이블은 soft delete(비활성) 우선
- 각 테이블은 공통 레이아웃/템플릿을 재사용하고, 초보가 이해하기 쉬운 형태로 구현
- 폼 validation(필수 입력 등)과 에러 메시지 표시

3) 페이징 (중요)
- 목록 페이지에 페이징 구현
- “첫 페이지(<<), 이전(<), 페이지 번호(1 2 3 …), 다음(>), 마지막(>>)”
- 가운데 페이지 번호 클릭 가능
- 페이지가 많을 때는 현재 페이지 기준으로 적절히 windowing (예: 1 … 4 5 6 … 20)

4) 디자인/UI 가이드
- 관리자 페이지는 “글자체가 선명”하고 “넓고 큼직”한 레이아웃을 추천
- 폰트/행간/버튼/입력폼은 크게(가독성 우선), 테이블도 넓게
- CSS는 단순하게(초보 수정 쉬움). Bootstrap 또는 Tailwind 중 하나 선택 가능하지만, 지나치게 복잡하게 만들지 말 것.
- HTMX 사용: 리스트에서 삭제/수정 후 부분 갱신(가능하면) 적용

[구현 제약/권장]
- Go 프로젝트 구조는 초보가 수정하기 쉬운 방식으로:
  - cmd/server/main.go
  - internal/http (router/handlers/middleware)
  - internal/service (business logic)
  - internal/repo (DB access)
  - internal/view (templates, helpers)
  - internal/auth (session, password)
  - internal/pagination (pager utility)
  - web/templates, web/static(css/js)
- DB access는 sqlc 또는 database/sql + 쿼리 파일 중 하나. 초보가 보기 쉬운 쪽을 선택하고 이유를 설명.
- SQL 인젝션 방지: 반드시 prepared statement/parameter binding 사용
- 에러 처리/로그를 일관성 있게 구성
- 환경변수(.env)로 DB 설정을 받을 수 있게 하되, 위 DB 정보를 기본 예시로 제공

[산출물]
1) database.sql을 기반으로 한 테이블 목록 정리 + 각 테이블의 CRUD 화면 구성(필수 필드/표시 필드 추천)
2) 전체 라우팅 표 (예: /login, /logout, /admin/<table>, /admin/<table>/new, /admin/<table>/:id, /admin/<table>/:id/edit)
3) 구현 순서(작게 쪼개서): 로그인 → 공통 레이아웃 → 한 테이블 CRUD → 제너럴 CRUD로 확장
4) 실제로 동작하는 코드 골격을 제공:
   - main.go
   - router/middleware
   - auth(session, bcrypt)
   - repository 예시(한 테이블)
   - generic list pagination helper
   - templates(레이아웃 + list/form/detail)
   - HTMX로 삭제/리스트 갱신 예시
5) 페이징 UI를 포함한 관리자 페이지 HTML/CSS 예시(가독성/큰 폰트/넓은 레이아웃)

[중요]
- database.sql 파일을 반드시 읽고 스키마에 맞춰 구현 방식을 제시해라.
- 모든 테이블 CRUD를 “복붙 지옥” 없이 확장 가능하게(제너릭/공통화) 하되, 초보가 이해 가능한 수준으로만 추상화해라.
- PostgreSQL MCP를 사용해 DB 스키마 확인/쿼리 실행/테이블 목록 확인을 수행해라.
[권한 정책 구현(추가)]
- internal/policy 기반으로 role + resource + action 권한 체크를 추가한다.
- 역할 정의: INTERNAL_SUPER_ADMIN(모든 권한), admin(규격/업체/BL포지션/차량번호 CRUD + 사용자 read), staff(입출고/BL마킹/휴가신청서 read/create, 본인 데이터 update/delete), supplier(업체 전용 조회 페이지 read).
- 각 페이지 핸들러에서 읽기/생성/수정/삭제 권한을 확인하고, staff는 소유자(user_id) 일치 시에만 수정/삭제를 허용한다.
- UI 네비게이션/버튼은 canAccess 헬퍼로 권한 있는 메뉴/액션만 노출한다.
- supplier 계정은 로그인 후 /supplier/portal 로 이동한다.
- policy_permissions 테이블을 추가하고 /admin/policies UI에서 역할별 CRUD 권한을 체크/저장할 수 있다.
