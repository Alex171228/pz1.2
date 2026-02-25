package service

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Done        bool   `json:"done"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

type TaskService struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewTaskService() *TaskService {
	return &TaskService{
		tasks: make(map[string]*Task),
	}
}

func (s *TaskService) Create(req CreateTaskRequest) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		ID:          "t_" + uuid.New().String()[:8],
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Done:        false,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	s.tasks[task.ID] = task
	return task
}

func (s *TaskService) GetAll() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *TaskService) GetByID(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *TaskService) Update(id string, req UpdateTaskRequest) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}
	if req.Done != nil {
		task.Done = *req.Done
	}

	return task, nil
}

func (s *TaskService) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return ErrTaskNotFound
	}

	delete(s.tasks, id)
	return nil
}
