package usecase

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/supchaser/test_task/internal/app"
	"github.com/supchaser/test_task/internal/app/models"
	"github.com/supchaser/test_task/internal/utils/logger"
	"go.uber.org/zap"
)

type TaskUsecase struct {
	taskRepository app.TaskRepository
	storagePath    string
}

func CreateTaskUsecase(taskRepository app.TaskRepository, storagePath string) *TaskUsecase {
	if storagePath == "" {
		storagePath = "./storage"
	}
	return &TaskUsecase{
		taskRepository: taskRepository,
		storagePath:    storagePath,
	}
}

func (u *TaskUsecase) CreateTask(ctx context.Context) (*models.Task, error) {
	const funcName = "TaskUsecase.CreateTask"
	logger.Debug("creating new task",
		zap.String("function", funcName),
	)

	task, err := u.taskRepository.CreateTask(ctx)
	if err != nil {
		logger.Error("failed to create task",
			zap.String("function", funcName),
			zap.Error(err),
		)
		return nil, err
	}

	return task, nil
}

func (u *TaskUsecase) GetTask(ctx context.Context, id int64) (*models.Task, error) {
	const funcName = "TaskUsecase.GetTask"
	logger.Debug("getting task",
		zap.String("function", funcName),
		zap.Int64("task_id", id),
	)

	task, err := u.taskRepository.GetTask(ctx, id)
	if err != nil {
		logger.Error("failed to get task",
			zap.String("function", funcName),
			zap.Int64("task_id", id),
			zap.Error(err),
		)
		return nil, err
	}

	return task, nil
}

func (u *TaskUsecase) AddObject(ctx context.Context, taskID int64, url string) (*models.Task, error) {
	const funcName = "TaskUsecase.AddObject"
	logger.Debug("adding object to task",
		zap.String("function", funcName),
		zap.Int64("task_id", taskID),
		zap.String("url", url),
	)

	task, err := u.taskRepository.AddObject(ctx, taskID, url)
	if err != nil {
		logger.Error("failed to add object",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, err
	}

	if len(task.Objects) == 3 {
		go u.ProcessTask(ctx, task.ID)
	}

	return task, nil
}

func (u *TaskUsecase) ProcessTask(ctx context.Context, taskID int64) {
	const funcName = "TaskUsecase.processTask"
	logger.Info("starting task processing",
		zap.String("function", funcName),
		zap.Int64("task_id", taskID),
	)

	if err := u.taskRepository.UpdateTaskStatus(ctx, taskID, models.StatusProcessing); err != nil {
		logger.Error("failed to update task status",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.Error(err),
		)
		return
	}

	task, err := u.taskRepository.GetTask(ctx, taskID)
	if err != nil {
		logger.Error("failed to get task for processing",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.Error(err),
		)
		return
	}

	zipPath := filepath.Join(u.storagePath, fmt.Sprintf("task_%d.zip", taskID))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		logger.Error("failed to create zip file",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.String("zip_path", zipPath),
			zap.Error(err),
		)
		u.taskRepository.UpdateTaskStatus(ctx, taskID, models.StatusFailed)
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	successCount := 0
	for _, obj := range task.Objects {
		resp, err := http.Get(obj.URL)
		if err != nil {
			logger.Warn("failed to download file",
				zap.String("function", funcName),
				zap.Int64("task_id", taskID),
				zap.String("url", obj.URL),
				zap.Error(err),
			)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warn("invalid response status",
				zap.String("function", funcName),
				zap.Int64("task_id", taskID),
				zap.String("url", obj.URL),
				zap.Int("status_code", resp.StatusCode),
			)
			continue
		}

		fileName := filepath.Base(obj.URL)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			logger.Warn("failed to create file in archive",
				zap.String("function", funcName),
				zap.Int64("task_id", taskID),
				zap.String("file_name", fileName),
				zap.Error(err),
			)
			continue
		}

		if _, err := io.Copy(fileWriter, resp.Body); err != nil {
			logger.Warn("failed to write file to archive",
				zap.String("function", funcName),
				zap.Int64("task_id", taskID),
				zap.String("file_name", fileName),
				zap.Error(err),
			)
			continue
		}

		successCount++
	}

	if successCount == 0 {
		logger.Error("no files were added to archive",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
		)
		u.taskRepository.UpdateTaskStatus(ctx, taskID, models.StatusFailed)
		os.Remove(zipPath)
		return
	}

	if err := u.taskRepository.UpdateTaskStatus(ctx, taskID, models.StatusDone); err != nil {
		logger.Error("failed to update task status",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.Error(err),
		)
		return
	}

	logger.Info("task processed successfully",
		zap.String("function", funcName),
		zap.Int64("task_id", taskID),
		zap.Int("files_processed", successCount),
		zap.Int("total_files", len(task.Objects)),
		zap.String("zip_path", zipPath),
	)
}

func (u *TaskUsecase) GetTaskStatus(ctx context.Context, id int64) (*models.Task, error) {
	const funcName = "TaskUsecase.GetTaskStatus"
	logger.Debug("getting task status",
		zap.String("function", funcName),
		zap.Int64("task_id", id),
	)

	task, err := u.taskRepository.GetTask(ctx, id)
	if err != nil {
		logger.Error("failed to get task status",
			zap.String("function", funcName),
			zap.Int64("task_id", id),
			zap.Error(err),
		)
		return nil, err
	}

	return task, nil
}

func (u *TaskUsecase) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	const funcName = "TaskUsecase.GetAllTasks"
	logger.Debug("getting all tasks",
		zap.String("function", funcName),
	)

	tasks, err := u.taskRepository.GetAllTasks(ctx)
	if err != nil {
		logger.Error("failed to get all tasks",
			zap.String("function", funcName),
			zap.Error(err),
		)
		return nil, err
	}

	return tasks, nil
}

func (u *TaskUsecase) GetMaxTasks() int {
	return u.taskRepository.GetMaxTasks()
}

func (u *TaskUsecase) GetActiveTasksCount() int {
	return u.taskRepository.GetActiveTasksCount()
}
