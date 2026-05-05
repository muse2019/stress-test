package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"stress-test/internal/report"
	"stress-test/pkg/models"
)

// ReportStore 报告存储接口
type ReportStore interface {
	List() ([]*models.Report, error)
	Get(id string) (*models.Report, error)
	Save(report *models.Report) error
	Delete(id string) error
	GetMarkdown(id string) (string, error)
}

// JSONReportStore JSON 报告存储实现
type JSONReportStore struct {
	dir string
}

// NewJSONReportStore 创建报告存储
func NewJSONReportStore(dir string) (*JSONReportStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &JSONReportStore{dir: dir}, nil
}

// List 获取报告列表（按时间倒序）
func (s *JSONReportStore) List() ([]*models.Report, error) {
	files, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var reports []*models.Report
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(file.Name(), ".json")
		r, err := s.Get(id)
		if err != nil {
			continue
		}
		reports = append(reports, r)
	}

	// 按开始时间倒序
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].StartTime.After(reports[j].StartTime)
	})

	return reports, nil
}

// Get 获取单个报告
func (s *JSONReportStore) Get(id string) (*models.Report, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}

	path := filepath.Join(s.dir, filepath.Base(id)+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("report not found")
		}
		return nil, err
	}

	var r models.Report
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

// Save 保存报告
func (s *JSONReportStore) Save(r *models.Report) error {
	if r == nil {
		return errors.New("report cannot be nil")
	}
	if r.ID == "" {
		r.ID = fmt.Sprintf("report-%d", time.Now().UnixNano())
	}
	if err := validateID(r.ID); err != nil {
		return err
	}

	// 保存 JSON
	path := filepath.Join(s.dir, filepath.Base(r.ID)+".json")
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// 保存 Markdown
	mdPath := filepath.Join(s.dir, filepath.Base(r.ID)+".md")
	md := report.GenerateMarkdown(r)
	return os.WriteFile(mdPath, []byte(md), 0644)
}

// Delete 删除报告
func (s *JSONReportStore) Delete(id string) error {
	if err := validateID(id); err != nil {
		return err
	}

	jsonPath := filepath.Join(s.dir, filepath.Base(id)+".json")
	mdPath := filepath.Join(s.dir, filepath.Base(id)+".md")

	os.Remove(mdPath)
	return os.Remove(jsonPath)
}

// GetMarkdown 获取 Markdown 内容
func (s *JSONReportStore) GetMarkdown(id string) (string, error) {
	if err := validateID(id); err != nil {
		return "", err
	}

	path := filepath.Join(s.dir, filepath.Base(id)+".md")

	data, err := os.ReadFile(path)
	if err != nil {
		// 如果 MD 文件不存在，从 JSON 生成
		r, err := s.Get(id)
		if err != nil {
			return "", err
		}
		return report.GenerateMarkdown(r), nil
	}

	return string(data), nil
}
