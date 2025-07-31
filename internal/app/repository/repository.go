package repository

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/supchaser/test_task/internal/app/models"
	"github.com/supchaser/test_task/internal/utils/errs"
	"github.com/supchaser/test_task/internal/utils/logger"
	"github.com/supchaser/test_task/internal/utils/validate"
	"go.uber.org/zap"
)

type TaskRepository struct {
	tasks       map[int64]*models.Task
	activeTasks int
	maxTasks    int
	mu          sync.Mutex
}

func CreateTaskRepository(maxTasks int) *TaskRepository {
	return &TaskRepository{
		tasks:    make(map[int64]*models.Task),
		maxTasks: maxTasks,
	}
}

func (r *TaskRepository) CreateTask(ctx context.Context) (*models.Task, error) {
	const funcName = "TaskRepository.CreateTask"
	logger.Debug("attempting to create task",
		zap.String("function", funcName),
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeTasks >= r.maxTasks {
		logger.Warn("maximum tasks limit reached",
			zap.String("function", funcName),
			zap.Int("active_tasks", r.activeTasks),
			zap.Int("max_tasks", r.maxTasks),
		)
		return nil, fmt.Errorf("%w: current %d, max %d", errs.ErrMaxTasksReached, r.activeTasks, r.maxTasks)
	}

	task := &models.Task{
		ID:        time.Now().UnixNano(),
		Status:    models.StatusWaiting,
		Objects:   make([]*models.Object, 0),
		CreatedAt: time.Now(),
	}

	r.tasks[task.ID] = task
	r.activeTasks++

	logger.Info("task created successfully",
		zap.String("function", funcName),
		zap.Int64("task_id", task.ID),
		zap.Int("active_tasks", r.activeTasks),
		zap.Time("created_at", task.CreatedAt),
	)

	return task, nil
}

func (r *TaskRepository) GetTask(ctx context.Context, id int64) (*models.Task, error) {
	const funcName = "TaskRepository.GetTask"
	logger.Debug("attempting to get task",
		zap.String("function", funcName),
		zap.Int64("task_id", id),
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[id]
	if !exists {
		logger.Warn("task not found",
			zap.String("function", funcName),
			zap.Int64("task_id", id),
		)
		return nil, errs.ErrTaskNotFound
	}

	logger.Info("task retrieved successfully",
		zap.String("function", funcName),
		zap.Int64("task_id", id),
		zap.String("status", string(task.Status)),
		zap.Int("objects_count", len(task.Objects)),
	)

	return task, nil
}

func (r *TaskRepository) AddObject(ctx context.Context, taskID int64, url string) (*models.Task, error) {
	const funcName = "TaskRepository.AddObject"
	logger.Debug("attempting to add object to task",
		zap.String("function", funcName),
		zap.Int64("task_id", taskID),
		zap.String("url", url),
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[taskID]
	if !exists {
		logger.Warn("task not found when adding object",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
		)
		return nil, errs.ErrTaskNotFound
	}

	if err := validate.ValidateObjectLimit(len(task.Objects)); err != nil {
		logger.Warn("maximum objects limit reached",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.Int("current_objects", len(task.Objects)),
			zap.Error(err),
		)
		return nil, err
	}

	if err := validate.ValidateFileExtension(url); err != nil {
		ext := strings.ToLower(filepath.Ext(url))
		logger.Warn("invalid file type",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.String("url", url),
			zap.String("extension", ext),
			zap.Error(err),
		)
		return nil, err
	}

	object := &models.Object{
		ID:  time.Now().UnixNano(),
		URL: url,
	}
	task.Objects = append(task.Objects, object)

	logger.Info("object added successfully",
		zap.String("function", funcName),
		zap.Int64("task_id", taskID),
		zap.String("url", url),
		zap.Int("new_objects_count", len(task.Objects)),
	)

	return task, nil
}

func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, id int64, status models.TaskStatus) error {
	const funcName = "TaskRepository.UpdateTaskStatus"
	logger.Debug("attempting to update task status",
		zap.String("function", funcName),
		zap.Int64("task_id", id),
		zap.String("new_status", string(status)),
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[id]
	if !exists {
		logger.Warn("task not found when updating status",
			zap.String("function", funcName),
			zap.Int64("task_id", id),
		)
		return errs.ErrTaskNotFound
	}

	oldStatus := task.Status
	task.Status = status

	if (status == models.StatusDone || status == models.StatusFailed) &&
		(oldStatus == models.StatusWaiting || oldStatus == models.StatusProcessing) {
		r.activeTasks--
		logger.Info("active task slot released",
			zap.String("function", funcName),
			zap.Int64("task_id", id),
			zap.Int("remaining_active_tasks", r.activeTasks),
		)
	}

	logger.Info("task status updated successfully",
		zap.String("function", funcName),
		zap.Int64("task_id", id),
		zap.String("old_status", string(oldStatus)),
		zap.String("new_status", string(status)),
	)

	return nil
}

func (r *TaskRepository) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	const funcName = "TaskRepository.GetAllTasks"
	logger.Debug("getting all tasks",
		zap.String("function", funcName),
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	tasks := make([]*models.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	logger.Info("retrieved all tasks",
		zap.String("function", funcName),
		zap.Int("count", len(tasks)),
	)

	return tasks, nil
}

func (r *TaskRepository) GetMaxTasks() int {
	return r.maxTasks
}

func (r *TaskRepository) GetActiveTasksCount() int {
	return r.activeTasks
}
