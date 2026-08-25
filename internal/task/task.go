package task

import (
	"sort"
	"time"

	"github.com/google/uuid"

	"greenhouse/internal/audit"
	"greenhouse/internal/store"
)

type Task struct {
	ID     string    `json:"id"`
	ShedID string    `json:"shed_id"`
	Name   string    `json:"name"`
	Kind   string    `json:"kind"`
	DueDay int       `json:"due_day"`
	Done   bool      `json:"done"`
	At     time.Time `json:"at"`
}

type TaskService struct {
	kv    *store.KeyValue
	audit *audit.Service
}

func NewTaskService(st *store.Store, audit *audit.Service) *TaskService {
	return &TaskService{kv: store.NewKeyValue(st), audit: audit}
}

func (t *TaskService) Create(task Task) (Task, error) {
	task.ID = uuid.NewString()
	task.At = time.Now()
	if err := t.kv.Save("tasks", task.ID, task); err != nil {
		return Task{}, err
	}
	if _, err := t.audit.Record(task.ShedID, "task", "create", task.Name); err != nil {
		return Task{}, err
	}
	return task, nil
}

func (t *TaskService) Complete(id string) error {
	var task Task
	if err := t.kv.Load("tasks", id, &task); err != nil {
		return err
	}
	task.Done = true
	if err := t.kv.Save("tasks", id, task); err != nil {
		return err
	}
	_, err := t.audit.Record(task.ShedID, "task", "complete", task.Name)
	return err
}

func (t *TaskService) List(shedID string, includeDone bool) ([]Task, error) {
	ids, err := t.kv.List("tasks")
	if err != nil {
		return nil, err
	}
	out := []Task{}
	for _, id := range ids {
		var task Task
		if err := t.kv.Load("tasks", id, &task); err != nil {
			continue
		}
		if task.ShedID == shedID && (includeDone || !task.Done) {
			out = append(out, task)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DueDay < out[j].DueDay })
	return out, nil
}

func itoa(value int64) string {
	return formatInt(value)
}
