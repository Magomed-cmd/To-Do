package export

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"todoapp/services/task-service/internal/adapters/http/middleware"
	"todoapp/services/task-service/internal/domain/entities"
	"todoapp/services/task-service/internal/ports"
)

// mockTaskService is a test double for ports.TaskService
type mockTaskService struct {
	exportData     []byte
	exportFilename string
	exportErr      error
	exportCalled   bool
	exportUserID   int64
	exportFormat   entities.ExportFormat
}

func (m *mockTaskService) ExportTasks(_ context.Context, userID int64, format entities.ExportFormat) ([]byte, string, error) {
	m.exportCalled = true
	m.exportUserID = userID
	m.exportFormat = format
	return m.exportData, m.exportFilename, m.exportErr
}

// Stub implementations for other interface methods
func (m *mockTaskService) CreateTask(_ context.Context, _ ports.CreateTaskInput) (*entities.Task, error) {
	return nil, nil
}
func (m *mockTaskService) UpdateTask(_ context.Context, _ ports.UpdateTaskInput) (*entities.Task, error) {
	return nil, nil
}
func (m *mockTaskService) UpdateTaskStatus(_ context.Context, _, _ int64, _ entities.TaskStatus) (*entities.Task, error) {
	return nil, nil
}
func (m *mockTaskService) DeleteTask(_ context.Context, _, _ int64) error                        { return nil }
func (m *mockTaskService) GetTask(_ context.Context, _, _ int64) (*entities.Task, error)         { return nil, nil }
func (m *mockTaskService) ListTasks(_ context.Context, _ int64, _ ports.TaskFilter) ([]entities.Task, error) {
	return nil, nil
}
func (m *mockTaskService) CreateCategory(_ context.Context, _ ports.CreateCategoryInput) (*entities.Category, error) {
	return nil, nil
}
func (m *mockTaskService) ListCategories(_ context.Context, _ int64) ([]entities.Category, error) {
	return nil, nil
}
func (m *mockTaskService) DeleteCategory(_ context.Context, _, _ int64) error { return nil }
func (m *mockTaskService) AddComment(_ context.Context, _ ports.AddCommentInput) (*entities.TaskComment, error) {
	return nil, nil
}
func (m *mockTaskService) ListComments(_ context.Context, _, _ int64) ([]entities.TaskComment, error) {
	return nil, nil
}

func setupTestRouter(service ports.TaskService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware that sets user claims
	router.Use(func(c *gin.Context) {
		claims := &ports.TokenClaims{
			UserID: 42,
			Email:  "test@example.com",
		}
		c.Set(middleware.ContextUserClaimsKey, claims)
		c.Next()
	})

	handler := New(service)
	handler.RegisterRoutes(router)

	return router
}

func setupTestRouterWithoutAuth(service ports.TaskService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := New(service)
	handler.RegisterRoutes(router)

	return router
}

func TestExportCSV_Success(t *testing.T) {
	mock := &mockTaskService{
		exportData:     []byte("ID,Title\n1,Test Task\n"),
		exportFilename: "tasks_2024-12-10.csv",
	}
	router := setupTestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/export/csv", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !mock.exportCalled {
		t.Error("expected ExportTasks to be called")
	}

	if mock.exportFormat != entities.ExportFormatCSV {
		t.Errorf("expected CSV format, got %s", mock.exportFormat)
	}

	if mock.exportUserID != 42 {
		t.Errorf("expected userID 42, got %d", mock.exportUserID)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("expected CSV content type, got %s", contentType)
	}

	disposition := w.Header().Get("Content-Disposition")
	if disposition != `attachment; filename="tasks_2024-12-10.csv"` {
		t.Errorf("unexpected Content-Disposition: %s", disposition)
	}
}

func TestExportICal_Success(t *testing.T) {
	mock := &mockTaskService{
		exportData:     []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"),
		exportFilename: "tasks_2024-12-10.ics",
	}
	router := setupTestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/export/ical", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if mock.exportFormat != entities.ExportFormatICal {
		t.Errorf("expected iCal format, got %s", mock.exportFormat)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/calendar; charset=utf-8" {
		t.Errorf("expected iCal content type, got %s", contentType)
	}
}

func TestExport_Unauthorized(t *testing.T) {
	mock := &mockTaskService{}
	router := setupTestRouterWithoutAuth(mock)

	req := httptest.NewRequest(http.MethodGet, "/export/csv", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	if mock.exportCalled {
		t.Error("ExportTasks should not be called for unauthorized request")
	}
}

func TestExport_ServiceError(t *testing.T) {
	mock := &mockTaskService{
		exportErr: errors.New("database connection failed"),
	}
	router := setupTestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/export/csv", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	mock := &mockTaskService{
		exportData:     []byte("test"),
		exportFilename: "test.csv",
	}
	router := setupTestRouter(mock)

	// Test CSV route exists
	req := httptest.NewRequest(http.MethodGet, "/export/csv", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Error("expected /export/csv route to exist")
	}

	// Test iCal route exists
	req = httptest.NewRequest(http.MethodGet, "/export/ical", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Error("expected /export/ical route to exist")
	}
}
