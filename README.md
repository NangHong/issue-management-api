# Go Issue Management API

이 프로젝트는 간단한 **이슈 관리 백엔드 API**를 Go 언어로 구현한 예제입니다.  
담당자 지정, 상태 변경, 필터링 기능 등을 포함한 CRUD API입니다.

---

## 📦 실행 환경

- Go 1.20 이상  
- 포트: `8080`

---

## 🚀 실행 방법

### 1. Go 설치

Go가 설치되어 있지 않다면 [https://golang.org/dl/](https://golang.org/dl/) 에서 운영체제에 맞는 설치 파일을 다운로드해 설치하세요.

설치 확인: cmd 실행 후

go version

###2. 프로젝트 클론

git clone https://github.com/NangHonh/issue-management-api.git
cd issue-management-api

###3. 의존성 설치

go mod tidy

###4. 서버 실행

go run main.go

###5 서버 확인
브라우저 또는 API 도구(Postman, curl 등)로 다음 주소에 접속해 확인합니다: http://localhost:8080

예) 이슈목록 확인: 
GET http://localhost:8080/issues

 API 목록
1. 이슈 생성 POST /issue
{
  "title": "버그 수정 필요",
  "description": "로그인 페이지에서 오류 발생",
  "userId": 1
}

2. 이슈 목록 조회
GET /issues
쿼리 파라미터: status (선택)

예: GET /issues?status=PENDING

3. 이슈 상세 조회
GET /issue/{id}

4. 이슈 수정
PATCH /issue/{id}

{
  "title": "로그인 버그 수정",
  "status": "IN_PROGRESS",
  "userId": 2
}

사전 등록 사용자

  { "id": 1, "name": "김개발" },
  { "id": 2, "name": "이디자인" },
  { "id": 3, "name": "박기획" }

에러 응답 형식
모든 에러는 다음과 같은 형식으로 응답됩니다:
{
  "error": "에러 메시지",
  "code": 400
}

 주요 라이브러리
Gorilla Mux: 라우팅 처리

net/http: Go 기본 HTTP 서버

GitHub
Repository: github.com/NangHonh/issue-management-api

작성자
GitHub: NangHonh

언어: Go (Golang)
