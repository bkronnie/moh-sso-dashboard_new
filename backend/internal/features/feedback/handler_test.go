package feedback

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/config"
)

func newTestHandler() *Handler {
	repo := NewHTTPRepository(&config.Config{}) // unconfigured on purpose
	return NewHandler(NewService(repo))
}

func TestFeedbackRoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	RegisterProtectedRoutes(router.Group(""), newTestHandler())

	want := map[string]bool{
		http.MethodGet + " /feedback":              false,
		http.MethodPatch + " /feedback/:id/triage": false,
	}
	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for route, found := range want {
		if !found {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}

func TestHandler_List_UnconfiguredReturns503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := newTestHandler()
	router.GET("/feedback", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/feedback", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusServiceUnavailable, "FEEDBACK_PROXY_NOT_CONFIGURED")
}

func TestHandler_UpdateTriage_InvalidIDRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := newTestHandler()
	router.PATCH("/feedback/:id/triage", handler.UpdateTriage)

	req := httptest.NewRequest(http.MethodPatch, "/feedback/not-a-number/triage", bytes.NewReader([]byte(`{"status":"New"}`)))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusBadRequest, "INVALID_ID")
}

func TestHandler_UpdateTriage_InvalidPayloadRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := newTestHandler()
	router.PATCH("/feedback/:id/triage", handler.UpdateTriage)

	req := httptest.NewRequest(http.MethodPatch, "/feedback/1/triage", bytes.NewReader([]byte(`{`)))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusBadRequest, "INVALID_PAYLOAD")
}

func TestHandler_UpdateTriage_InvalidStatusRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := newTestHandler()
	router.PATCH("/feedback/:id/triage", handler.UpdateTriage)

	req := httptest.NewRequest(http.MethodPatch, "/feedback/1/triage", bytes.NewReader([]byte(`{"status":"Not A Real Status"}`)))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusBadRequest, "INVALID_STATUS")
}

func assertErrorCode(t *testing.T, res *httptest.ResponseRecorder, status int, code string) {
	t.Helper()

	if res.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, res.Code, res.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Success {
		t.Fatal("expected failure response")
	}
	if body.Error.Code != code {
		t.Fatalf("expected error code %q, got %q", code, body.Error.Code)
	}
}
