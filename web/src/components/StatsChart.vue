<script setup lang="ts">
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
  ToolboxComponent,
} from 'echarts/components'
import type { RealtimeStats } from '@/types'

// Register ECharts components
use([
  CanvasRenderer,
  LineChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
  ToolboxComponent,
])

interface Props {
  timeline: RealtimeStats[]
  title?: string
  height?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  height: '500px',
})

// Shared color palette
const colors = {
  qps: '#6366f1',
  avgRT: '#22c55e',
  minRT: '#14b8a6',
  maxRT: '#f59e0b',
  successRate: '#10b981',
  failed: '#ef4444',
}

// Format time for x-axis
const times = computed(() =>
  props.timeline.map(s =>
    new Date(s.timestamp).toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  )
)

// QPS & Success Rate chart option
const qpsChartOption = computed(() => {
  const qps = props.timeline.map(s => s.qps.toFixed(2))
  const successRates = props.timeline.map(s => {
    const total = s.totalRequests
    if (total === 0) return 100
    return ((s.successCount / total) * 100).toFixed(2)
  })

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
        crossStyle: { color: '#999' },
      },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e5e7eb',
      borderWidth: 1,
      textStyle: { color: '#374151' },
      formatter: (params: any) => {
        let result = `<div style="font-weight:600;margin-bottom:8px">${params[0].axisValue}</div>`
        params.forEach((item: any) => {
          const unit = item.seriesName === '成功率' ? '%' : ''
          result += `<div style="display:flex;align-items:center;gap:8px;margin:4px 0">
            <span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${item.color}"></span>
            <span>${item.seriesName}: <strong>${item.value}${unit}</strong></span>
          </div>`
        })
        return result
      },
    },
    legend: {
      data: ['QPS', '成功率'],
      top: 10,
      textStyle: { color: '#6b7280' },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '15%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: times.value,
      boundaryGap: false,
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: {
        color: '#9ca3af',
        fontSize: 11,
        rotate: 30,
      },
      axisTick: { show: false },
    },
    yAxis: [
      {
        type: 'value',
        name: 'QPS',
        nameTextStyle: { color: '#9ca3af', fontSize: 12 },
        position: 'left',
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: { color: '#9ca3af' },
        splitLine: { lineStyle: { color: '#f3f4f6', type: 'dashed' } },
      },
      {
        type: 'value',
        name: '成功率 (%)',
        nameTextStyle: { color: '#9ca3af', fontSize: 12 },
        position: 'right',
        min: 0,
        max: 100,
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: { color: '#9ca3af', formatter: '{value}%' },
        splitLine: { show: false },
      },
    ],
    series: [
      {
        name: 'QPS',
        type: 'line',
        smooth: 0.6,
        symbol: 'circle',
        symbolSize: 6,
        showSymbol: false,
        lineStyle: { width: 3, color: colors.qps },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(99, 102, 241, 0.3)' },
              { offset: 1, color: 'rgba(99, 102, 241, 0.02)' },
            ],
          },
        },
        emphasis: {
          focus: 'series',
          itemStyle: { shadowBlur: 10, shadowColor: 'rgba(99, 102, 241, 0.5)' },
        },
        data: qps,
      },
      {
        name: '成功率',
        type: 'line',
        smooth: 0.6,
        symbol: 'circle',
        symbolSize: 6,
        showSymbol: false,
        yAxisIndex: 1,
        lineStyle: { width: 2, color: colors.successRate, type: 'dashed' },
        itemStyle: { color: colors.successRate },
        data: successRates,
      },
    ],
  }
})

// Response Time chart option
const rtChartOption = computed(() => {
  const avgRT = props.timeline.map(s => s.avgRT)
  const minRT = props.timeline.map(s => s.minRT)
  const maxRT = props.timeline.map(s => s.maxRT)

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e5e7eb',
      borderWidth: 1,
      textStyle: { color: '#374151' },
      formatter: (params: any) => {
        let result = `<div style="font-weight:600;margin-bottom:8px">${params[0].axisValue}</div>`
        params.forEach((item: any) => {
          result += `<div style="display:flex;align-items:center;gap:8px;margin:4px 0">
            <span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${item.color}"></span>
            <span>${item.seriesName}: <strong>${item.value}ms</strong></span>
          </div>`
        })
        return result
      },
    },
    legend: {
      data: ['平均响应时间', '最小RT', '最大RT'],
      top: 10,
      textStyle: { color: '#6b7280' },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '15%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: times.value,
      boundaryGap: false,
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: {
        color: '#9ca3af',
        fontSize: 11,
        rotate: 30,
      },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      name: '响应时间 (ms)',
      nameTextStyle: { color: '#9ca3af', fontSize: 12 },
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#9ca3af' },
      splitLine: { lineStyle: { color: '#f3f4f6', type: 'dashed' } },
    },
    series: [
      {
        name: '平均响应时间',
        type: 'line',
        smooth: 0.6,
        symbol: 'circle',
        symbolSize: 6,
        showSymbol: false,
        lineStyle: { width: 3, color: colors.avgRT },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(34, 197, 94, 0.25)' },
              { offset: 1, color: 'rgba(34, 197, 94, 0.02)' },
            ],
          },
        },
        emphasis: {
          focus: 'series',
          itemStyle: { shadowBlur: 10, shadowColor: 'rgba(34, 197, 94, 0.5)' },
        },
        data: avgRT,
      },
      {
        name: '最小RT',
        type: 'line',
        smooth: 0.6,
        symbol: 'none',
        lineStyle: { width: 1, color: colors.minRT, type: 'dotted' },
        itemStyle: { color: colors.minRT },
        data: minRT,
      },
      {
        name: '最大RT',
        type: 'line',
        smooth: 0.6,
        symbol: 'none',
        lineStyle: { width: 1, color: colors.maxRT, type: 'dotted' },
        itemStyle: { color: colors.maxRT },
        data: maxRT,
      },
    ],
  }
})

// Percentile distribution chart option (bar chart for last stats)
const percentileChartOption = computed(() => {
  if (props.timeline.length === 0) return null

  const lastStats = props.timeline[props.timeline.length - 1]
  const percentiles = ['P50', 'P90', 'P95', 'P99']
  const values = [lastStats.p50, lastStats.p90, lastStats.p95, lastStats.p99]

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e5e7eb',
      borderWidth: 1,
      textStyle: { color: '#374151' },
      formatter: (params: any) => {
        return `<div style="font-weight:600">${params[0].name}</div>
                <div style="margin-top:4px">${params[0].value}ms</div>`
      },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: percentiles,
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: { color: '#6b7280', fontSize: 13, fontWeight: 500 },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      name: 'ms',
      nameTextStyle: { color: '#9ca3af' },
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#9ca3af' },
      splitLine: { lineStyle: { color: '#f3f4f6', type: 'dashed' } },
    },
    series: [
      {
        type: 'bar',
        barWidth: '50%',
        data: values.map((value, index) => ({
          value,
          itemStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: ['#818cf8', '#a78bfa', '#c084fc', '#e879f9'][index] },
                { offset: 1, color: ['#6366f1', '#8b5cf6', '#a855f7', '#d946ef'][index] },
              ],
            },
            borderRadius: [6, 6, 0, 0],
          },
        })),
        label: {
          show: true,
          position: 'top',
          formatter: '{c}ms',
          color: '#6b7280',
          fontSize: 12,
          fontWeight: 500,
        },
        emphasis: {
          itemStyle: {
            shadowBlur: 20,
            shadowColor: 'rgba(99, 102, 241, 0.3)',
          },
        },
      },
    ],
  }
})
</script>

<template>
  <div class="charts-container">
    <div class="chart-section" v-if="timeline.length > 0">
      <div class="chart-title">
        <span class="title-icon qps"></span>
        QPS 与成功率趋势
      </div>
      <v-chart
        :option="qpsChartOption"
        class="chart"
        autoresize
      />
    </div>

    <div class="chart-section" v-if="timeline.length > 0">
      <div class="chart-title">
        <span class="title-icon rt"></span>
        响应时间趋势
      </div>
      <v-chart
        :option="rtChartOption"
        class="chart"
        autoresize
      />
    </div>

    <div class="chart-section percentile-section" v-if="percentileChartOption">
      <div class="chart-title">
        <span class="title-icon percentile"></span>
        响应时间分布 (百分位)
      </div>
      <v-chart
        :option="percentileChartOption"
        class="chart percentile-chart"
        autoresize
      />
    </div>
  </div>
</template>

<style scoped>
.charts-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.chart-section {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.1);
}

.chart-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 16px;
}

.title-icon {
  width: 4px;
  height: 18px;
  border-radius: 2px;
}

.title-icon.qps {
  background: linear-gradient(180deg, #6366f1, #8b5cf6);
}

.title-icon.rt {
  background: linear-gradient(180deg, #22c55e, #14b8a6);
}

.title-icon.percentile {
  background: linear-gradient(180deg, #818cf8, #d946ef);
}

.chart {
  height: 320px;
  width: 100%;
}

.percentile-chart {
  height: 280px;
}

.percentile-section {
  max-width: 500px;
}
</style>
