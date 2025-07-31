package delivery

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	mock_app "github.com/supchaser/test_task/internal/app/mocks"
	"github.com/supchaser/test_task/internal/app/models"
	"github.com/supchaser/test_task/internal/utils/errs"
	"github.com/supchaser/test_task/internal/utils/logger"
)

func TestMain(m *testing.M) {
	logger.InitTestLogger()
	m.Run()
}

func TestTaskDelivery_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockTaskUsecase(ctrl)
	taskDelivery := CreateTaskDelivery(mockUsecase)

	tests := []struct {
		name             string
		mockSetup        func()
		expectedStatus   int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockUsecase.EXPECT().
					CreateTask(gomock.Any()).
					Return(&models.Task{
						ID:        1,
						Status:    models.StatusWaiting,
						CreatedAt: time.Now(),
						Objects:   []*models.Object{},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var task models.Task
				err := json.Unmarshal(body, &task)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), task.ID)
				assert.Equal(t, models.StatusWaiting, task.Status)
				assert.NotZero(t, task.CreatedAt)
				assert.Empty(t, task.Objects)
			},
		},
		{
			name: "MaxTasksReached",
			mockSetup: func() {
				mockUsecase.EXPECT().
					CreateTask(gomock.Any()).
					Return(nil, errs.ErrMaxTasksReached)
				mockUsecase.EXPECT().
					GetMaxTasks().
					Return(5)
				mockUsecase.EXPECT().
					GetActiveTasksCount().
					Return(5)
			},
			expectedStatus: http.StatusTooManyRequests,
			validateResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, errs.ErrMaxTasksReached.Error(), response["error"])
				assert.Equal(t, float64(5), response["max_tasks"])
				assert.Equal(t, float64(5), response["active_now"])
				assert.Contains(t, response["suggestion"], "Try again later")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("POST", "/tasks", nil)
			w := httptest.NewRecorder()

			taskDelivery.CreateTask(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w.Body.Bytes())
		})
	}
}

func TestTaskDelivery_GetTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockTaskUsecase(ctrl)
	taskDelivery := CreateTaskDelivery(mockUsecase)

	tests := []struct {
		name           string
		taskID         string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:   "Success",
			taskID: "1",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(&models.Task{ID: 1}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "InvalidID",
			taskID:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "TaskNotFound",
			taskID: "1",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(nil, errs.ErrTaskNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()

			vars := map[string]string{
				"id": tt.taskID,
			}
			req = mux.SetURLVars(req, vars)

			taskDelivery.GetTask(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestTaskDelivery_AddObjects(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockTaskUsecase(ctrl)
	taskDelivery := CreateTaskDelivery(mockUsecase)

	tests := []struct {
		name           string
		taskID         string
		requestBody    any
		mockSetup      func(*mock_app.MockTaskUsecase)
		expectedStatus int
		expectedResult *models.MultiAddResult
	}{
		{
			name:   "SuccessSingleURL",
			taskID: "1",
			requestBody: map[string][]string{
				"urls": {"http://example.com/image.jpg"},
			},
			mockSetup: func(m *mock_app.MockTaskUsecase) {
				m.EXPECT().
					AddObject(gomock.Any(), int64(1), "http://example.com/image.jpg").
					Return(&models.Task{ID: 1, Objects: []*models.Object{{URL: "http://example.com/image.jpg"}}}, nil)
				m.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(&models.Task{ID: 1, Objects: []*models.Object{{URL: "http://example.com/image.jpg"}}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResult: &models.MultiAddResult{
				AddedCount:   1,
				FailedURLs:   map[string]string{},
				TotalObjects: 1,
			},
		},
		{
			name:   "SuccessMultipleURLs",
			taskID: "1",
			requestBody: map[string][]string{
				"urls": {
					"http://example.com/image1.jpg",
					"http://example.com/image2.jpg",
				},
			},
			mockSetup: func(m *mock_app.MockTaskUsecase) {
				m.EXPECT().
					AddObject(gomock.Any(), int64(1), "http://example.com/image1.jpg").
					Return(&models.Task{ID: 1}, nil)
				m.EXPECT().
					AddObject(gomock.Any(), int64(1), "http://example.com/image2.jpg").
					Return(&models.Task{ID: 1}, nil)
				m.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(&models.Task{
						ID: 1,
						Objects: []*models.Object{
							{URL: "http://example.com/image1.jpg"},
							{URL: "http://example.com/image2.jpg"},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResult: &models.MultiAddResult{
				AddedCount:   2,
				FailedURLs:   map[string]string{},
				TotalObjects: 2,
			},
		},
		{
			name:   "InvalidTaskID",
			taskID: "invalid",
			requestBody: map[string][]string{
				"urls": {"http://example.com/image.jpg"},
			},
			mockSetup:      func(m *mock_app.MockTaskUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "InvalidRequestBody_InvalidJSON",
			taskID:         "1",
			requestBody:    "invalid json",
			mockSetup:      func(m *mock_app.MockTaskUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "TaskNotFound",
			taskID: "1",
			requestBody: map[string][]string{
				"urls": {"http://example.com/image.jpg"},
			},
			mockSetup: func(m *mock_app.MockTaskUsecase) {
				m.EXPECT().
					AddObject(gomock.Any(), int64(1), "http://example.com/image.jpg").
					Return(nil, errs.ErrTaskNotFound)
				m.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(nil, errs.ErrTaskNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "PartialSuccessWithFailures",
			taskID: "1",
			requestBody: map[string][]string{
				"urls": {
					"http://example.com/good.jpg",
					"http://example.com/bad.jpg",
				},
			},
			mockSetup: func(m *mock_app.MockTaskUsecase) {
				m.EXPECT().
					AddObject(gomock.Any(), int64(1), "http://example.com/good.jpg").
					Return(&models.Task{ID: 1}, nil)
				m.EXPECT().
					AddObject(gomock.Any(), int64(1), "http://example.com/bad.jpg").
					Return(nil, errs.ErrInvalidFileType)
				m.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(&models.Task{
						ID: 1,
						Objects: []*models.Object{
							{URL: "http://example.com/good.jpg"},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResult: &models.MultiAddResult{
				AddedCount: 1,
				FailedURLs: map[string]string{
					"http://example.com/bad.jpg": "invalid file type (allowed: .pdf, .jpeg)",
				},
				TotalObjects: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockUsecase)

			var body []byte
			switch v := tt.requestBody.(type) {
			case string:
				body = []byte(v)
			default:
				var err error
				body, err = json.Marshal(v)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/tasks/"+tt.taskID+"/objects", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			vars := map[string]string{
				"id": tt.taskID,
			}
			req = mux.SetURLVars(req, vars)

			taskDelivery.AddObjects(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var result models.MultiAddResult
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.AddedCount, result.AddedCount)
				assert.Equal(t, tt.expectedResult.TotalObjects, result.TotalObjects)
				assert.Equal(t, len(tt.expectedResult.FailedURLs), len(result.FailedURLs))

				for url, errMsg := range tt.expectedResult.FailedURLs {
					assert.Contains(t, result.FailedURLs, url)
					assert.Equal(t, errMsg, result.FailedURLs[url])
				}
			}
		})
	}
}

func TestTaskDelivery_GetTaskStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockTaskUsecase(ctrl)
	taskDelivery := CreateTaskDelivery(mockUsecase)

	tests := []struct {
		name           string
		taskID         string
		mockSetup      func()
		expectedStatus int
		expectedZipURL bool
	}{
		{
			name:   "Success_Waiting",
			taskID: "1",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetTaskStatus(gomock.Any(), int64(1)).
					Return(&models.Task{ID: 1, Status: models.StatusWaiting}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedZipURL: false,
		},
		{
			name:   "Success_Done",
			taskID: "1",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetTaskStatus(gomock.Any(), int64(1)).
					Return(&models.Task{ID: 1, Status: models.StatusDone}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedZipURL: true,
		},
		{
			name:           "InvalidTaskID",
			taskID:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/tasks/"+tt.taskID+"/status", nil)
			w := httptest.NewRecorder()

			vars := map[string]string{
				"id": tt.taskID,
			}
			req = mux.SetURLVars(req, vars)

			taskDelivery.GetTaskStatus(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Status string `json:"status"`
					ZipURL string `json:"zip_url,omitempty"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				if tt.expectedZipURL {
					assert.NotEmpty(t, response.ZipURL)
				} else {
					assert.Empty(t, response.ZipURL)
				}
			}
		})
	}
}

func TestTaskDelivery_DownloadArchive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockTaskUsecase(ctrl)
	taskDelivery := CreateTaskDelivery(mockUsecase)

	tests := []struct {
		name           string
		taskID         string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:           "InvalidTaskID",
			taskID:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "TaskNotFound",
			taskID: "1",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(nil, errs.ErrTaskNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "ArchiveNotReady",
			taskID: "1",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(&models.Task{Status: models.StatusProcessing}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/download/"+tt.taskID, nil)
			w := httptest.NewRecorder()

			vars := map[string]string{
				"id": tt.taskID,
			}
			req = mux.SetURLVars(req, vars)

			taskDelivery.DownloadArchive(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestTaskDelivery_GetAllTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockTaskUsecase(ctrl)
	taskDelivery := CreateTaskDelivery(mockUsecase)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "SuccessWithTasks",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetAllTasks(gomock.Any()).
					Return([]*models.Task{
						{ID: 1},
						{ID: 2},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "SuccessNoTasks",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetAllTasks(gomock.Any()).
					Return([]*models.Task{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/tasks", nil)
			w := httptest.NewRecorder()

			taskDelivery.GetAllTasks(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedCount == 0 {
				assert.Equal(t, "No tasks found", response["message"])
			} else {
				tasks := response["tasks"].([]interface{})
				assert.Equal(t, tt.expectedCount, len(tasks))
			}
		})
	}
}
