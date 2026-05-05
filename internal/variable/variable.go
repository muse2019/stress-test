package variable

import (
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"

	"stress-test/pkg/models"
)

// Processor 变量处理器
type Processor struct {
	variables map[string]*VariableValue
	csvData   map[string][]string // CSV 列数据
	csvIndex  map[string]int       // CSV 当前行索引
}

// VariableValue 变量值
type VariableValue struct {
	varType string
	value   string
	min     int
	max     int
}

// NewProcessor 创建变量处理器
func NewProcessor(vars []models.Variable) (*Processor, error) {
	p := &Processor{
		variables: make(map[string]*VariableValue),
		csvData:   make(map[string][]string),
		csvIndex:  make(map[string]int),
	}

	for _, v := range vars {
		switch v.Type {
		case "static":
			p.variables[v.Name] = &VariableValue{
				varType: "static",
				value:   v.Value,
			}
		case "random_int":
			p.variables[v.Name] = &VariableValue{
				varType: "random_int",
				min:     v.Min,
				max:     v.Max,
			}
		case "random_string":
			p.variables[v.Name] = &VariableValue{
				varType: "random_string",
				min:     v.Min, // length
			}
		case "uuid":
			p.variables[v.Name] = &VariableValue{
				varType: "uuid",
			}
		case "csv":
			if err := p.loadCSV(v.Name, v.File, v.Column); err != nil {
				return nil, fmt.Errorf("failed to load CSV for %s: %w", v.Name, err)
			}
		}
	}

	return p, nil
}

// loadCSV 加载 CSV 文件
func (p *Processor) loadCSV(name, file, column string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return fmt.Errorf("empty CSV file")
	}

	// 找到列索引
	header := records[0]
	colIndex := -1
	for i, h := range header {
		if h == column {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		// 尝试解析为数字索引
		if idx, err := strconv.Atoi(column); err == nil && idx < len(header) {
			colIndex = idx
		} else {
			return fmt.Errorf("column %s not found", column)
		}
	}

	// 提取列数据
	var values []string
	for i := 1; i < len(records); i++ {
		if colIndex < len(records[i]) {
			values = append(values, records[i][colIndex])
		}
	}

	p.csvData[name] = values
	p.csvIndex[name] = 0
	return nil
}

// Get 获取变量值
func (p *Processor) Get(name string) (string, error) {
	v, ok := p.variables[name]
	if ok {
		return p.generateValue(v)
	}

	// 检查 CSV 数据
	if data, ok := p.csvData[name]; ok {
		if len(data) == 0 {
			return "", fmt.Errorf("empty CSV data for %s", name)
		}
		idx := p.csvIndex[name] % len(data)
		value := data[idx]
		p.csvIndex[name]++
		return value, nil
	}

	return "", fmt.Errorf("variable %s not found", name)
}

// generateValue 生成变量值
func (p *Processor) generateValue(v *VariableValue) (string, error) {
	switch v.varType {
	case "static":
		return v.value, nil

	case "random_int":
		min := v.min
		max := v.max
		if max <= min {
			max = min + 100
		}
		n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
		if err != nil {
			return "", err
		}
		return strconv.Itoa(int(n.Int64()) + min), nil

	case "random_string":
		length := v.min
		if length <= 0 {
			length = 10
		}
		const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		result := make([]byte, length)
		for i := range result {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
			result[i] = chars[n.Int64()]
		}
		return string(result), nil

	case "uuid":
		return generateUUID(), nil

	default:
		return "", fmt.Errorf("unknown variable type: %s", v.varType)
	}
}

// generateUUID 生成 UUID
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Replace 替换字符串中的变量
func (p *Processor) Replace(s string) (string, error) {
	// 匹配 {{varName}} 格式
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)

	var err error
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		if err != nil {
			return match
		}
		name := strings.Trim(match, "{}")
		value, e := p.Get(name)
		if e != nil {
			err = e
			return match
		}
		return value
	})

	return result, err
}

// ReplaceMap 替换 map 中的变量
func (p *Processor) ReplaceMap(m map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for k, v := range m {
		replaced, err := p.Replace(v)
		if err != nil {
			return nil, err
		}
		result[k] = replaced
	}
	return result, nil
}
