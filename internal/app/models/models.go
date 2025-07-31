package models

import "time"

type TaskStatus string

const (
	StatusWaiting    TaskStatus = "waiting"
	StatusProcessing TaskStatus = "processing"
	StatusDone       TaskStatus = "done"
	StatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID        int64
	Status    TaskStatus
	Objects   []*Object
	CreatedAt time.Time
}

type Object struct {
	ID    int64
	URL   string
	Error string
}

type Request struct {
	URLs []string `json:"urls"`
}

type TaskResponse struct {
	ID           int64      `json:"id"`
	Status       TaskStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ObjectsCount int        `json:"objects_count"`
}

type MultiAddResult struct {
	AddedCount   int               `json:"added_count"`
	FailedURLs   map[string]string `json:"failed_urls"`
	TotalObjects int               `json:"total_objects"`
}
