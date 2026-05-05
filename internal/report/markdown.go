package report

import (
	"fmt"
	"strings"

	"stress-test/pkg/models"
)

// GenerateMarkdown 生成 Markdown 报告
func GenerateMarkdown(report *models.Report) string {
	var sb strings.Builder

	// 标题
	sb.WriteString(fmt.Sprintf("# 压测报告: %s\n\n", report.TaskName))

	// 基本信息
	sb.WriteString("## 基本信息\n\n")
	sb.WriteString("| 项目 | 值 |\n")
	sb.WriteString("|------|-----|\n")
	sb.WriteString(fmt.Sprintf("| 任务名称 | %s |\n", report.TaskName))
	sb.WriteString(fmt.Sprintf("| 协议 | HTTP |\n"))
	sb.WriteString(fmt.Sprintf("| 目标地址 | %s |\n", report.Config.Target))
	sb.WriteString(fmt.Sprintf("| 请求方法 | %s |\n", report.Config.Method))
	sb.WriteString(fmt.Sprintf("| 压测模式 | %s |\n", getModeText(report.Config.Mode)))
	sb.WriteString(fmt.Sprintf("| 并发数 | %d |\n", report.Config.Concurrency))
	sb.WriteString(fmt.Sprintf("| 持续时间 | %ds |\n", report.Duration))
	sb.WriteString(fmt.Sprintf("| 开始时间 | %s |\n", report.StartTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("| 结束时间 | %s |\n\n", report.EndTime.Format("2006-01-02 15:04:05")))

	// 性能指标
	sb.WriteString("## 性能指标\n\n")
	sb.WriteString("| 指标 | 值 |\n")
	sb.WriteString("|------|-----|\n")
	sb.WriteString(fmt.Sprintf("| 总请求数 | %d |\n", report.FinalStats.TotalRequests))
	sb.WriteString(fmt.Sprintf("| 成功数 | %d |\n", report.FinalStats.SuccessCount))
	sb.WriteString(fmt.Sprintf("| 失败数 | %d |\n", report.FinalStats.FailedCount))
	sb.WriteString(fmt.Sprintf("| 成功率 | %.2f%% |\n", report.FinalStats.SuccessRate()))
	sb.WriteString(fmt.Sprintf("| QPS | %.1f |\n", report.FinalStats.QPS))
	sb.WriteString(fmt.Sprintf("| 平均响应时间 | %dms |\n", report.FinalStats.AvgRT))
	sb.WriteString(fmt.Sprintf("| 最小响应时间 | %dms |\n", report.FinalStats.MinRT))
	sb.WriteString(fmt.Sprintf("| 最大响应时间 | %dms |\n\n", report.FinalStats.MaxRT))

	// 响应时间分布
	sb.WriteString("## 响应时间分布\n\n")
	sb.WriteString("| 分位数 | 响应时间 |\n")
	sb.WriteString("|--------|----------|\n")
	sb.WriteString(fmt.Sprintf("| P50 | %dms |\n", report.FinalStats.P50))
	sb.WriteString(fmt.Sprintf("| P90 | %dms |\n", report.FinalStats.P90))
	sb.WriteString(fmt.Sprintf("| P95 | %dms |\n", report.FinalStats.P95))
	sb.WriteString(fmt.Sprintf("| P99 | %dms |\n\n", report.FinalStats.P99))

	// 错误分布
	if len(report.FinalStats.Errors) > 0 {
		sb.WriteString("## 错误分布\n\n")
		sb.WriteString("| 错误类型 | 次数 |\n")
		sb.WriteString("|----------|------|\n")
		for errType, count := range report.FinalStats.Errors {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", errType, count))
		}
		sb.WriteString("\n")
	}

	// 页脚
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("*报告生成时间: %s*\n", report.EndTime.Format("2006-01-02 15:04:05")))

	return sb.String()
}

func getModeText(mode string) string {
	switch mode {
	case "fixed":
		return "固定并发"
	case "staircase":
		return "阶梯递增"
	case "rate":
		return "QPS限制"
	default:
		return mode
	}
}
