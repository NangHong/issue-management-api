package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "sync"
    "time"

    "github.com/gorilla/mux" // HTTP 라우팅 라이브러리
)

// User 구조체: 사용자 정보 저장용
type User struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

// Issue 구조체: 이슈 정보 저장용
type Issue struct {
    ID          uint      `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    User        *User     `json:"user,omitempty"` // 담당자 정보 (없으면 생략)
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

var (
    // 미리 정의된 사용자 3명 (ID:1,2,3)
    users = map[uint]*User{
        1: {ID: 1, Name: "김개발"},
        2: {ID: 2, Name: "이디자인"},
        3: {ID: 3, Name: "박기획"},
    }

    issues     = map[uint]*Issue{} // 이슈 저장소 (메모리)
    issueIDSeq uint = 1            // 이슈 ID 자동 증가 시퀀스

    mu sync.Mutex // 동시성 보호용 뮤텍스
)

// 유효한 이슈 상태값 집합
var validStatuses = map[string]bool{
    "PENDING":     true,
    "IN_PROGRESS": true,
    "COMPLETED":   true,
    "CANCELLED":   true,
}

// 상태값 유효성 검사 함수
func isValidStatus(s string) bool {
    return validStatuses[s]
}

// 에러 응답 공통 함수
func respondWithError(w http.ResponseWriter, code int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": message,
        "code":  code,
    })
}

// [POST] /issue - 이슈 생성 핸들러
func createIssue(w http.ResponseWriter, r *http.Request) {
    type createIssueRequest struct {
        Title       string  `json:"title"`
        Description *string `json:"description,omitempty"`
        UserID      *uint   `json:"userId,omitempty"`
    }

    // 요청 바디 JSON 파싱
    var req createIssueRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
        return
    }

    // 필수 필드 검사: title
    if req.Title == "" {
        respondWithError(w, http.StatusBadRequest, "필수 필드 title 누락")
        return
    }

    var user *User
    status := "PENDING" // 기본 상태

    // 담당자(userId) 있으면 유효성 검사 후 상태 변경
    if req.UserID != nil {
        mu.Lock()
        u, ok := users[*req.UserID]
        mu.Unlock()
        if !ok {
            respondWithError(w, http.StatusBadRequest, "유효하지 않은 userId")
            return
        }
        user = u
        status = "IN_PROGRESS"
    }

    now := time.Now()

    mu.Lock()
    issue := &Issue{
        ID:          issueIDSeq,
        Title:       req.Title,
        Description: "",
        Status:      status,
        User:        user,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
    if req.Description != nil {
        issue.Description = *req.Description
    }
    issues[issueIDSeq] = issue
    issueIDSeq++
    mu.Unlock()

    // 응답: 201 Created + 생성된 이슈 JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(issue)
}

// [GET] /issues - 이슈 목록 조회 핸들러 (상태 필터 가능)
func getAllIssues(w http.ResponseWriter, r *http.Request) {
    statusFilter := r.URL.Query().Get("status")

    // status 파라미터가 있다면 유효성 검사
    if statusFilter != "" && !isValidStatus(statusFilter) {
        respondWithError(w, http.StatusBadRequest, "유효하지 않은 status")
        return
    }

    mu.Lock()
    defer mu.Unlock()

    var filtered []*Issue
    for _, issue := range issues {
        if statusFilter == "" || issue.Status == statusFilter {
            filtered = append(filtered, issue)
        }
    }

    // 응답 형식 맞춰서 JSON 출력
    response := struct {
        Issues []*Issue `json:"issues"`
    }{
        Issues: filtered,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// [GET] /issue/{id} - 이슈 상세 조회 핸들러
func getIssueByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        respondWithError(w, http.StatusBadRequest, "Invalid issue id")
        return
    }

    mu.Lock()
    issue, ok := issues[uint(id)]
    mu.Unlock()

    if !ok {
        respondWithError(w, http.StatusNotFound, "Issue not found")
        return
    }

    // 상세 응답에는 담당자 정보 제외 (요구사항에 맞춤)
    resp := struct {
        ID          uint      `json:"id"`
        Title       string    `json:"title"`
        Description string    `json:"description"`
        Status      string    `json:"status"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`
    }{
        ID:          issue.ID,
        Title:       issue.Title,
        Description: issue.Description,
        Status:      issue.Status,
        CreatedAt:   issue.CreatedAt,
        UpdatedAt:   issue.UpdatedAt,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

// [PATCH] /issue/{id} - 이슈 수정 핸들러
func updateIssue(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        respondWithError(w, http.StatusBadRequest, "Invalid issue id")
        return
    }

    type updateIssueRequest struct {
        Title       *string `json:"title,omitempty"`
        Description *string `json:"description,omitempty"`
        Status      *string `json:"status,omitempty"`
        UserID      *uint   `json:"userId,omitempty"` // nil: 변경 없음, 0: 담당자 제거
    }

    var req updateIssueRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
        return
    }

    mu.Lock()
    issue, ok := issues[uint(id)]
    mu.Unlock()
    if !ok {
        respondWithError(w, http.StatusNotFound, "Issue not found")
        return
    }

    // 완료/취소 상태 이슈는 수정 불가
    if issue.Status == "COMPLETED" || issue.Status == "CANCELLED" {
        respondWithError(w, http.StatusForbidden, "Completed or cancelled issues cannot be updated")
        return
    }

    mu.Lock()
    defer mu.Unlock()

    user := issue.User // 현재 담당자 정보

    if req.UserID != nil {
        if *req.UserID == 0 {
            // 담당자 제거 + 상태 PENDING
            user = nil
            issue.Status = "PENDING"
        } else {
            u, exists := users[*req.UserID]
            if !exists {
                respondWithError(w, http.StatusBadRequest, "유효하지 않은 userId")
                return
            }
            user = u

            // 담당자 할당되고 상태가 PENDING일 때 상태 변경
            if issue.Status == "PENDING" && (req.Status == nil) {
                issue.Status = "IN_PROGRESS"
            }
        }
    }

    if req.Status != nil {
        if !isValidStatus(*req.Status) {
            respondWithError(w, http.StatusBadRequest, "유효하지 않은 status")
            return
        }
        // 담당자 없는데 PENDING, CANCELLED 외 상태 지정 불가
        if user == nil && *req.Status != "PENDING" && *req.Status != "CANCELLED" {
            respondWithError(w, http.StatusBadRequest, "담당자 없는 상태에서 PENDING 또는 CANCELLED 상태만 지정 가능")
            return
        }
        issue.Status = *req.Status
    }

    if req.Title != nil {
        issue.Title = *req.Title
    }
    if req.Description != nil {
        issue.Description = *req.Description
    }

    issue.User = user
    issue.UpdatedAt = time.Now()

    // 수정된 이슈 반환
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(issue)
}

func main() {
    r := mux.NewRouter() // Gorilla mux 라우터 생성

    // 라우팅 경로 및 메서드 연결
    r.HandleFunc("/issue", createIssue).Methods("POST")
    r.HandleFunc("/issues", getAllIssues).Methods("GET")
    r.HandleFunc("/issue/{id}", getIssueByID).Methods("GET")
    r.HandleFunc("/issue/{id}", updateIssue).Methods("PATCH")

    // 8080 포트로 서버 실행
    log.Println("Starting server at :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
