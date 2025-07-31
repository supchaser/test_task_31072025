package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/supchaser/test_task/internal/app"
	"github.com/supchaser/test_task/internal/app/models"
	"github.com/supchaser/test_task/internal/utils/errs"
	"github.com/supchaser/test_task/internal/utils/logger"
	"github.com/supchaser/test_task/internal/utils/responses"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type TaskDelivery struct {
	taskUsecase app.TaskUsecase
}

func CreateTaskDelivery(taskUsecase app.TaskUsecase) *TaskDelivery {
	return &TaskDelivery{
		taskUsecase: taskUsecase,
	}
}

func (d *TaskDelivery) CreateTask(w http.ResponseWriter, r *http.Request) {
	const funcName = "TaskDelivery.CreateTask"
	logger.Debug("creating new task", zap.String("function", funcName))

	task, err := d.taskUsecase.CreateTask(r.Context())
	if err != nil {
		if errors.Is(err, errs.ErrMaxTasksReached) {
			responses.DoJSONResponse(w, map[string]any{
				"error":      err.Error(),
				"max_tasks":  d.taskUsecase.GetMaxTasks(),
				"active_now": d.taskUsecase.GetActiveTasksCount(),
				"suggestion": "Try again later or wait for current tasks to complete",
			}, http.StatusTooManyRequests)
			return
		}
		responses.ResponseErrorAndLog(w, err, funcName)
		return
	}

	responses.DoJSONResponse(w, task, http.StatusCreated)
}

func (d *TaskDelivery) GetTask(w http.ResponseWriter, r *http.Request) {
	const funcName = "TaskDelivery.GetTask"
	logger.Debug("getting task",
		zap.String("function", funcName),
	)

	vars := mux.Vars(r)
	rawID := vars["id"]
	taskID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		responses.DoBadResponseAndLog(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := d.taskUsecase.GetTask(r.Context(), taskID)
	if err != nil {
		responses.ResponseErrorAndLog(w, err, funcName)
		return
	}

	responses.DoJSONResponse(w, task, http.StatusOK)
}

func (d *TaskDelivery) AddObjects(w http.ResponseWriter, r *http.Request) {
	const funcName = "TaskDelivery.AddObjects"
	logger.Debug("adding multiple objects to task",
		zap.String("function", funcName),
	)

	vars := mux.Vars(r)
	taskID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		responses.DoBadResponseAndLog(w, http.StatusBadRequest, "invalid task id")
		return
	}

	req := models.Request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.DoBadResponseAndLog(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.URLs) > 3 {
		responses.DoBadResponseAndLog(w, http.StatusBadRequest, "maximum 3 urls per request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result := &models.MultiAddResult{
		AddedCount:   0,
		FailedURLs:   make(map[string]string),
		TotalObjects: 0,
	}

	mu := sync.Mutex{}
	g, ctx := errgroup.WithContext(ctx)
	for _, url := range req.URLs {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			_, err := d.taskUsecase.AddObject(ctx, taskID, url)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.FailedURLs[url] = err.Error()
				logger.Warn("failed to add object",
					zap.String("url", url),
					zap.Error(err),
				)
			} else {
				result.AddedCount++
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		responses.DoBadResponseAndLog(w, http.StatusInternalServerError, "processing error")
		return
	}

	task, err := d.taskUsecase.GetTask(r.Context(), taskID)
	if err != nil {
		responses.ResponseErrorAndLog(w, err, funcName)
		return
	}

	result.TotalObjects = len(task.Objects)
	responses.DoJSONResponse(w, result, http.StatusOK)
}

func (d *TaskDelivery) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	const funcName = "TaskDelivery.GetTaskStatus"
	logger.Debug("getting task status",
		zap.String("function", funcName),
	)

	vars := mux.Vars(r)
	rawID := vars["id"]
	taskID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		responses.DoBadResponseAndLog(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := d.taskUsecase.GetTaskStatus(r.Context(), taskID)
	if err != nil {
		responses.ResponseErrorAndLog(w, err, funcName)
		return
	}

	response := struct {
		Status models.TaskStatus `json:"status"`
		ZipURL string            `json:"zip_url,omitempty"`
		Errors []string          `json:"errors,omitempty"`
	}{
		Status: task.Status,
	}

	if task.Status == models.StatusDone {
		response.ZipURL = "/download/" + strconv.FormatInt(taskID, 10)
	}

	responses.DoJSONResponse(w, response, http.StatusOK)
}

func (d *TaskDelivery) DownloadArchive(w http.ResponseWriter, r *http.Request) {
	const funcName = "TaskDelivery.DownloadArchive"
	logger.Debug("downloading archive",
		zap.String("function", funcName),
	)

	vars := mux.Vars(r)
	rawID := vars["id"]
	taskID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		logger.Warn("invalid task id",
			zap.String("function", funcName),
			zap.Error(err),
		)
		responses.DoBadResponseAndLog(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := d.taskUsecase.GetTask(r.Context(), taskID)
	if err != nil {
		logger.Error("failed to get task",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.Error(err),
		)
		responses.ResponseErrorAndLog(w, err, funcName)
		return
	}

	if task.Status != models.StatusDone {
		logger.Warn("archive not ready",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.String("status", string(task.Status)),
		)
		responses.DoBadResponseAndLog(w, http.StatusNotFound, "archive not ready")
		return
	}

	zipPath := fmt.Sprintf("./storage/task_%d.zip", taskID)

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		logger.Error("archive file not found",
			zap.String("function", funcName),
			zap.Int64("task_id", taskID),
			zap.String("path", zipPath),
			zap.Error(err),
		)
		responses.DoBadResponseAndLog(w, http.StatusInternalServerError, "archive file missing")
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=task_%d.zip", taskID))

	http.ServeFile(w, r, zipPath)

	logger.Info("archive downloaded successfully",
		zap.String("function", funcName),
		zap.Int64("task_id", taskID),
	)
}

func (d *TaskDelivery) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	const funcName = "TaskDelivery.GetAllTasks"
	logger.Debug("getting all tasks",
		zap.String("function", funcName),
	)

	tasks, err := d.taskUsecase.GetAllTasks(r.Context())
	if err != nil {
		responses.ResponseErrorAndLog(w, err, funcName)
		return
	}

	if len(tasks) == 0 {
		responses.DoJSONResponse(w, map[string]any{
			"message":    "No tasks found",
			"suggestion": "Create a new task with POST /api/v1/tasks",
			"count":      0,
			"tasks":      []any{},
		}, http.StatusOK)
		return
	}

	response := make([]models.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, models.TaskResponse{
			ID:           task.ID,
			Status:       task.Status,
			CreatedAt:    task.CreatedAt,
			ObjectsCount: len(task.Objects),
		})
	}

	responses.DoJSONResponse(w, map[string]any{
		"count": len(response),
		"tasks": response,
	}, http.StatusOK)
}
