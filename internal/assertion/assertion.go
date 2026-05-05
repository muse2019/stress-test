package assertion

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"stress-test/internal/engine"
	"stress-test/pkg/models"
)

// Result 断言结果
type Result struct {
	Passed bool   `json:"passed"`
	Type   string `json:"type"`
	Rule   string `json:"rule"`
	Error  string `json:"error,omitempty"`
}

// Validator 断言验证器
type Validator struct {
	assertions []models.Assertion
}

// NewValidator 创建验证器
func NewValidator(assertions []models.Assertion) *Validator {
	return &Validator{assertions: assertions}
}

// Validate 验证请求结果
func (v *Validator) Validate(result *engine.Result) []Result {
	var results []Result

	for _, a := range v.assertions {
		r := v.validateAssertion(a, result)
		results = append(results, r)
	}

	return results
}

// validateAssertion 验证单个断言
func (v *Validator) validateAssertion(a models.Assertion, result *engine.Result) Result {
	r := Result{
		Type: a.Type,
		Rule: fmt.Sprintf("%s %s %v", a.Type, a.Operator, a.Expected),
	}

	switch a.Type {
	case "statusCode":
		r.Passed = v.validateStatusCode(a.Operator, a.Expected, result.StatusCode)
		if !r.Passed {
			r.Error = fmt.Sprintf("status code %d does not match expected %v", result.StatusCode, a.Expected)
		}

	case "responseTime":
		expectedMS := toInt(a.Expected)
		actualMS := result.Latency.Milliseconds()
		r.Passed = v.validateValue(a.Operator, float64(actualMS), float64(expectedMS))
		if !r.Passed {
			r.Error = fmt.Sprintf("response time %dms does not match expected %dms", actualMS, expectedMS)
		}

	case "body":
		bodyStr := string(result.Body)
		r.Passed = v.validateBody(a.Operator, a.Expected, bodyStr)
		if !r.Passed {
			r.Error = fmt.Sprintf("body does not match expected %v", a.Expected)
		}

	default:
		r.Passed = false
		r.Error = fmt.Sprintf("unknown assertion type: %s", a.Type)
	}

	return r
}

// validateStatusCode 验证状态码
func (v *Validator) validateStatusCode(operator string, expected interface{}, actual int) bool {
	expectedInt := toInt(expected)

	switch operator {
	case "eq":
		return actual == expectedInt
	case "ne":
		return actual != expectedInt
	default:
		return false
	}
}

// validateValue 验证数值
func (v *Validator) validateValue(operator string, actual, expected float64) bool {
	switch operator {
	case "eq":
		return actual == expected
	case "ne":
		return actual != expected
	case "lt":
		return actual < expected
	case "gt":
		return actual > expected
	case "lte":
		return actual <= expected
	case "gte":
		return actual >= expected
	default:
		return false
	}
}

// validateBody 验证响应体
func (v *Validator) validateBody(operator string, expected interface{}, body string) bool {
	expectedStr := fmt.Sprintf("%v", expected)

	switch operator {
	case "eq":
		return body == expectedStr
	case "ne":
		return body != expectedStr
	case "contains":
		return strings.Contains(body, expectedStr)
	case "notContains":
		return !strings.Contains(body, expectedStr)
	case "regex":
		matched, err := regexp.MatchString(expectedStr, body)
		if err != nil {
			return false
		}
		return matched
	default:
		return false
	}
}

// toInt 转换为整数
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

// AllPassed 检查所有断言是否通过
func AllPassed(results []Result) bool {
	for _, r := range results {
		if !r.Passed {
			return false
		}
	}
	return true
}

// HasFailures 检查是否有失败的断言
func HasFailures(results []Result) bool {
	return !AllPassed(results)
}

// Summary 断言摘要
type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
}

// GetSummary 获取断言摘要
func GetSummary(results []Result) Summary {
	s := Summary{Total: len(results)}
	for _, r := range results {
		if r.Passed {
			s.Passed++
		} else {
			s.Failed++
		}
	}
	return s
}
