package app

import (
	"context"

	"github.com/supchaser/test_task/internal/app/models"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go

type TaskRepository interface {
	CreateTask(ctx context.Context) (*models.Task, error)
	GetTask(ctx context.Context, id int64) (*models.Task, error)
	AddObject(ctx context.Context, taskID int64, url string) (*models.Task, error)
	UpdateTaskStatus(ctx context.Context, id int64, status models.TaskStatus) error
	GetAllTasks(ctx context.Context) ([]*models.Task, error)
	GetMaxTasks() int
	GetActiveTasksCount() int
}

type TaskUsecase interface {
	CreateTask(ctx context.Context) (*models.Task, error)
	GetTask(ctx context.Context, id int64) (*models.Task, error)
	AddObject(ctx context.Context, taskID int64, url string) (*models.Task, error)
	GetTaskStatus(ctx context.Context, id int64) (*models.Task, error)
	GetAllTasks(ctx context.Context) ([]*models.Task, error)
	GetMaxTasks() int
	GetActiveTasksCount() int
}
