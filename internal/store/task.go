package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"stress-test/pkg/models"
)

var (
	// ErrInvalidID is returned when task ID is empty or contains invalid characters
	ErrInvalidID = errors.New("invalid task id: must be non-empty and contain only alphanumeric, dash, or underscore characters")
	// ErrNilTask is returned when a nil task is passed to Save
	ErrNilTask = errors.New("task cannot be nil")
	// ErrTaskNotFound is returned when a task does not exist
	ErrTaskNotFound = errors.New("task not found")
)

// validIDRegex matches IDs containing only alphanumeric, dash, or underscore
var validIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// validateID checks if the ID is valid and safe for use in file paths
func validateID(id string) error {
	if id == "" {
		return ErrInvalidID
	}
	if !validIDRegex.MatchString(id) {
		return ErrInvalidID
	}
	return nil
}

// TaskStore 任务存储接口
type TaskStore interface {
	List() ([]*models.Task, error)
	Get(id string) (*models.Task, error)
	Save(task *models.Task) error
	Delete(id string) error
}

// JSONTaskStore JSON 文件存储实现
type JSONTaskStore struct {
	dir string
}

// NewJSONTaskStore 创建 JSON 任务存储
func NewJSONTaskStore(dir string) (*JSONTaskStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &JSONTaskStore{dir: dir}, nil
}

// List 获取所有任务
func (s *JSONTaskStore) List() ([]*models.Task, error) {
	files, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []*models.Task
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(file.Name(), ".json")
		task, err := s.Get(id)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Get 获取单个任务
func (s *JSONTaskStore) Get(id string) (*models.Task, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}

	// Use filepath.Base as additional protection against path traversal
	path := filepath.Join(s.dir, filepath.Base(id)+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// Save 保存任务
func (s *JSONTaskStore) Save(task *models.Task) error {
	if task == nil {
		return ErrNilTask
	}
	if err := validateID(task.ID); err != nil {
		return err
	}

	// Use filepath.Base as additional protection against path traversal
	path := filepath.Join(s.dir, filepath.Base(task.ID)+".json")

	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Delete 删除任务
func (s *JSONTaskStore) Delete(id string) error {
	if err := validateID(id); err != nil {
		return err
	}

	// Use filepath.Base as additional protection against path traversal
	path := filepath.Join(s.dir, filepath.Base(id)+".json")

	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrTaskNotFound
		}
		return err
	}
	return nil
}
