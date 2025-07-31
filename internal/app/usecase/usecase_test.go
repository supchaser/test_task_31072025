package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
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

func TestTaskUsecase_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func(*mock_app.MockTaskRepository)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					CreateTask(gomock.Any()).
					Return(&models.Task{
						ID:        1,
						Status:    models.StatusWaiting,
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedTask: &models.Task{
				ID:     1,
				Status: models.StatusWaiting,
			},
			expectedError: nil,
		},
		{
			name: "MaxTasksReached",
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					CreateTask(gomock.Any()).
					Return(nil, errs.ErrMaxTasksReached)
			},
			expectedTask:  nil,
			expectedError: errs.ErrMaxTasksReached,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock_app.NewMockTaskRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			uc := CreateTaskUsecase(mockRepo, "")
			result, err := uc.CreateTask(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, result.ID)
				assert.Equal(t, tt.expectedTask.Status, result.Status)
				assert.NotZero(t, result.CreatedAt)
			}
		})
	}
}

func TestTaskUsecase_GetTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		taskID        int64
		mockSetup     func(*mock_app.MockTaskRepository)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name:   "Success",
			taskID: 1,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(&models.Task{
						ID:     1,
						Status: models.StatusProcessing,
					}, nil)
			},
			expectedTask: &models.Task{
				ID:     1,
				Status: models.StatusProcessing,
			},
			expectedError: nil,
		},
		{
			name:   "TaskNotFound",
			taskID: 2,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					GetTask(gomock.Any(), int64(2)).
					Return(nil, errs.ErrTaskNotFound)
			},
			expectedTask:  nil,
			expectedError: errs.ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock_app.NewMockTaskRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			uc := CreateTaskUsecase(mockRepo, "")
			result, err := uc.GetTask(context.Background(), tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, result.ID)
				assert.Equal(t, tt.expectedTask.Status, result.Status)
			}
		})
	}
}

func TestTaskUsecase_AddObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validURL := "http://example.com/image.jpg"
	invalidURL := "http://example.com/document.docx"

	tests := []struct {
		name          string
		taskID        int64
		url           string
		mockSetup     func(*mock_app.MockTaskRepository)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name:   "Success",
			taskID: 1,
			url:    validURL,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					AddObject(gomock.Any(), int64(1), validURL).
					Return(&models.Task{
						ID:     1,
						Status: models.StatusWaiting,
						Objects: []*models.Object{
							{URL: validURL},
						},
					}, nil)
			},
			expectedTask: &models.Task{
				ID:     1,
				Status: models.StatusWaiting,
				Objects: []*models.Object{
					{URL: validURL},
				},
			},
			expectedError: nil,
		},
		{
			name:   "TaskNotFound",
			taskID: 2,
			url:    validURL,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					AddObject(gomock.Any(), int64(2), validURL).
					Return(nil, errs.ErrTaskNotFound)
			},
			expectedTask:  nil,
			expectedError: errs.ErrTaskNotFound,
		},
		{
			name:   "InvalidFileExtension",
			taskID: 1,
			url:    invalidURL,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					AddObject(gomock.Any(), int64(1), invalidURL).
					Return(nil, errs.ErrInvalidFileType)
			},
			expectedTask:  nil,
			expectedError: errs.ErrInvalidFileType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock_app.NewMockTaskRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			uc := CreateTaskUsecase(mockRepo, "")
			result, err := uc.AddObject(context.Background(), tt.taskID, tt.url)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, result.ID)
				assert.Equal(t, len(tt.expectedTask.Objects), len(result.Objects))
				if len(result.Objects) > 0 {
					assert.Equal(t, tt.expectedTask.Objects[0].URL, result.Objects[0].URL)
				}
			}
		})
	}
}

func TestTaskUsecase_GetTaskStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		taskID        int64
		mockSetup     func(*mock_app.MockTaskRepository)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name:   "Success",
			taskID: 1,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(&models.Task{
						ID:     1,
						Status: models.StatusDone,
					}, nil)
			},
			expectedTask: &models.Task{
				ID:     1,
				Status: models.StatusDone,
			},
			expectedError: nil,
		},
		{
			name:   "TaskNotFound",
			taskID: 2,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					GetTask(gomock.Any(), int64(2)).
					Return(nil, errs.ErrTaskNotFound)
			},
			expectedTask:  nil,
			expectedError: errs.ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock_app.NewMockTaskRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			uc := CreateTaskUsecase(mockRepo, "")
			result, err := uc.GetTaskStatus(context.Background(), tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, result.ID)
				assert.Equal(t, tt.expectedTask.Status, result.Status)
			}
		})
	}
}

func TestTaskUsecase_GetAllTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func(*mock_app.MockTaskRepository)
		expectedTasks []*models.Task
		expectedError error
	}{
		{
			name: "SuccessWithTasks",
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					GetAllTasks(gomock.Any()).
					Return([]*models.Task{
						{ID: 1, Status: models.StatusWaiting},
						{ID: 2, Status: models.StatusProcessing},
					}, nil)
			},
			expectedTasks: []*models.Task{
				{ID: 1, Status: models.StatusWaiting},
				{ID: 2, Status: models.StatusProcessing},
			},
			expectedError: nil,
		},
		{
			name: "SuccessNoTasks",
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					GetAllTasks(gomock.Any()).
					Return([]*models.Task{}, nil)
			},
			expectedTasks: []*models.Task{},
			expectedError: nil,
		},
		{
			name: "RepositoryError",
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					GetAllTasks(gomock.Any()).
					Return(nil, assert.AnError)
			},
			expectedTasks: nil,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock_app.NewMockTaskRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			uc := CreateTaskUsecase(mockRepo, "")
			result, err := uc.GetAllTasks(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedTasks), len(result))
				for i, task := range tt.expectedTasks {
					assert.Equal(t, task.ID, result[i].ID)
					assert.Equal(t, task.Status, result[i].Status)
				}
			}
		})
	}
}

func TestTaskUsecase_processTask(t *testing.T) {
	tempDir := t.TempDir()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test file content"))
	}))
	defer testServer.Close()

	tests := []struct {
		name          string
		taskID        int64
		objects       []*models.Object
		mockSetup     func(*mock_app.MockTaskRepository)
		storagePath   string
		expectStatus  models.TaskStatus
		expectZipFile bool
	}{
		{
			name:   "SuccessProcessing",
			taskID: 1,
			objects: []*models.Object{
				{URL: testServer.URL + "/image1.jpg"},
				{URL: testServer.URL + "/image2.jpg"},
			},
			storagePath: tempDir,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), int64(1), models.StatusProcessing).
					Return(nil)

				mockRepo.EXPECT().
					GetTask(gomock.Any(), int64(1)).
					Return(&models.Task{
						ID:     1,
						Status: models.StatusProcessing,
						Objects: []*models.Object{
							{URL: testServer.URL + "/image1.jpg"},
							{URL: testServer.URL + "/image2.jpg"},
						},
					}, nil)

				mockRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), int64(1), models.StatusDone).
					Return(nil)
			},
			expectStatus:  models.StatusDone,
			expectZipFile: true,
		},
		{
			name:   "FailedToCreateZip_DirectoryNotWritable",
			taskID: 2,
			objects: []*models.Object{
				{URL: testServer.URL + "/image1.jpg"},
			},
			storagePath: "/non/existing/path",
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), int64(2), models.StatusProcessing).
					Return(nil)

				mockRepo.EXPECT().
					GetTask(gomock.Any(), int64(2)).
					Return(&models.Task{
						ID:     2,
						Status: models.StatusProcessing,
						Objects: []*models.Object{
							{URL: testServer.URL + "/image1.jpg"},
						},
					}, nil)

				mockRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), int64(2), models.StatusFailed).
					Return(nil)
			},
			expectStatus:  models.StatusFailed,
			expectZipFile: false,
		},
		{
			name:   "NoValidFiles",
			taskID: 3,
			objects: []*models.Object{
				{URL: "http://invalid.url/bad.docx"},
			},
			storagePath: tempDir,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), int64(3), models.StatusProcessing).
					Return(nil)

				mockRepo.EXPECT().
					GetTask(gomock.Any(), int64(3)).
					Return(&models.Task{
						ID:     3,
						Status: models.StatusProcessing,
						Objects: []*models.Object{
							{URL: "http://invalid.url/bad.docx"},
						},
					}, nil)

				mockRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), int64(3), models.StatusFailed).
					Return(nil)
			},
			expectStatus:  models.StatusFailed,
			expectZipFile: false,
		},
		{
			name:   "SuccessWithPDF",
			taskID: 4,
			objects: []*models.Object{
				{URL: testServer.URL + "/document.pdf"},
			},
			storagePath: tempDir,
			mockSetup: func(mockRepo *mock_app.MockTaskRepository) {
				mockRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), int64(4), models.StatusProcessing).
					Return(nil)

				mockRepo.EXPECT().
					GetTask(gomock.Any(), int64(4)).
					Return(&models.Task{
						ID:     4,
						Status: models.StatusProcessing,
						Objects: []*models.Object{
							{URL: testServer.URL + "/document.pdf"},
						},
					}, nil)

				mockRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), int64(4), models.StatusDone).
					Return(nil)
			},
			expectStatus:  models.StatusDone,
			expectZipFile: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock_app.NewMockTaskRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			uc := CreateTaskUsecase(mockRepo, tt.storagePath)
			uc.ProcessTask(context.Background(), tt.taskID)
			zipPath := filepath.Join(tt.storagePath, fmt.Sprintf("task_%d.zip", tt.taskID))
			if tt.expectZipFile {
				if _, err := os.Stat(zipPath); os.IsNotExist(err) {
					t.Errorf("expected zip file to be created at %s", zipPath)
				}
			} else {
				if _, err := os.Stat(zipPath); err == nil {
					t.Errorf("expected no zip file at %s", zipPath)
				}
			}
		})
	}
}

func TestTaskUsecase_GetMaxTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedMax := 5
	mockRepo := mock_app.NewMockTaskRepository(ctrl)
	mockRepo.EXPECT().GetMaxTasks().Return(expectedMax)

	uc := CreateTaskUsecase(mockRepo, "")
	result := uc.GetMaxTasks()

	assert.Equal(t, expectedMax, result)
}

func TestTaskUsecase_GetActiveTasksCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedCount := 3
	mockRepo := mock_app.NewMockTaskRepository(ctrl)
	mockRepo.EXPECT().GetActiveTasksCount().Return(expectedCount)

	uc := CreateTaskUsecase(mockRepo, "")
	result := uc.GetActiveTasksCount()

	assert.Equal(t, expectedCount, result)
}
