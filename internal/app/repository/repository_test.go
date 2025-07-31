package repository

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/supchaser/test_task/internal/app/models"
	"github.com/supchaser/test_task/internal/utils/errs"
	"github.com/supchaser/test_task/internal/utils/logger"
)

func TestMain(m *testing.M) {
	logger.InitTestLogger()
	m.Run()
}

func TestCreateTask_Success(t *testing.T) {
	repo := CreateTaskRepository(3)

	task, err := repo.CreateTask(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, models.StatusWaiting, task.Status)
	assert.Empty(t, task.Objects)
	assert.True(t, task.ID > 0)
	assert.WithinDuration(t, time.Now(), task.CreatedAt, time.Second)
}

func TestCreateTask_MaxTasksReached(t *testing.T) {
	maxTasks := 2
	repo := CreateTaskRepository(maxTasks)

	for range maxTasks {
		_, err := repo.CreateTask(context.Background())
		assert.NoError(t, err)
	}

	task, err := repo.CreateTask(context.Background())

	assert.Nil(t, task)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errs.ErrMaxTasksReached)
}

func TestGetTask_Success(t *testing.T) {
	repo := CreateTaskRepository(5)
	createdTask, err := repo.CreateTask(context.Background())
	assert.NoError(t, err)

	task, err := repo.GetTask(context.Background(), createdTask.ID)

	assert.NoError(t, err)
	assert.Equal(t, createdTask.ID, task.ID)
	assert.Equal(t, createdTask.Status, task.Status)
}

func TestGetTask_NotFound(t *testing.T) {
	repo := CreateTaskRepository(5)
	nonExistentID := int64(999999)

	task, err := repo.GetTask(context.Background(), nonExistentID)

	assert.Nil(t, task)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errs.ErrTaskNotFound)
}

func TestAddObject_Success(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	repo := CreateTaskRepository(5)
	createdTask, err := repo.CreateTask(context.Background())
	assert.NoError(t, err)

	validURL := testServer.URL + "/image.jpg"

	task, err := repo.AddObject(context.Background(), createdTask.ID, validURL)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, 1, len(task.Objects))
	assert.Equal(t, validURL, task.Objects[0].URL)
	assert.True(t, task.Objects[0].ID > 0)
}

func TestAddObject_InvalidExtension(t *testing.T) {
	repo := CreateTaskRepository(5)
	createdTask, err := repo.CreateTask(context.Background())
	assert.NoError(t, err)
	invalidURL := "http://example.com/document.docx"

	task, err := repo.AddObject(context.Background(), createdTask.ID, invalidURL)

	assert.Nil(t, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file type (allowed: .pdf, .jpeg)")
}

func TestAddObject_TaskNotFound(t *testing.T) {
	repo := CreateTaskRepository(5)
	nonExistentID := int64(999999)
	validURL := "http://example.com/image.jpg"

	task, err := repo.AddObject(context.Background(), nonExistentID, validURL)

	assert.Nil(t, task)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errs.ErrTaskNotFound)
}

func TestUpdateTaskStatus_Success(t *testing.T) {
	repo := CreateTaskRepository(5)
	createdTask, err := repo.CreateTask(context.Background())
	assert.NoError(t, err)

	err = repo.UpdateTaskStatus(context.Background(), createdTask.ID, models.StatusProcessing)

	assert.NoError(t, err)
	task, _ := repo.GetTask(context.Background(), createdTask.ID)
	assert.Equal(t, models.StatusProcessing, task.Status)
}

func TestUpdateTaskStatus_DecreasesActiveCount(t *testing.T) {
	repo := CreateTaskRepository(5)
	createdTask, err := repo.CreateTask(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, repo.GetActiveTasksCount())

	err = repo.UpdateTaskStatus(context.Background(), createdTask.ID, models.StatusDone)

	assert.NoError(t, err)
	assert.Equal(t, 0, repo.GetActiveTasksCount())
}

func TestGetAllTasks(t *testing.T) {
	repo := CreateTaskRepository(5)
	count := 3
	for range count {
		_, err := repo.CreateTask(context.Background())
		assert.NoError(t, err)
	}

	tasks, err := repo.GetAllTasks(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, count, len(tasks))
}

func TestGetAllTasks_Empty(t *testing.T) {
	repo := CreateTaskRepository(5)

	tasks, err := repo.GetAllTasks(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, tasks)
}
