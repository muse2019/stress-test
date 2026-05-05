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
])

interface Props {
  timeline: RealtimeStats[]
  title?: string
  height?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  height: '300px',
})

const chartOption = computed(() => {
  const times = props.timeline.map(s =>
    new Date(s.timestamp).toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  )
  const qps = props.timeline.map(s => s.qps)
  const avgRT = props.timeline.map(s => s.avgRT)
  const successRates = props.timeline.map(s => {
    const total = s.totalRequests
    if (total === 0) return 100
    return (s.successCount / total) * 100
  })

  return {
    title: props.title
      ? {
          text: props.title,
          left: 'center',
          textStyle: { fontSize: 14 },
        }
      : undefined,
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
    },
    legend: {
      data: ['QPS', 'Avg RT (ms)', 'Success Rate (%)'],
      bottom: 0,
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: times,
      boundaryGap: false,
      axisLabel: {
        rotate: 45,
        fontSize: 10,
      },
    },
    yAxis: [
      {
        type: 'value',
        name: 'QPS / RT',
        position: 'left',
      },
      {
        type: 'value',
        name: 'Success %',
        position: 'right',
        min: 0,
        max: 100,
      },
    ],
    series: [
      {
        name: 'QPS',
        type: 'line',
        smooth: true,
        data: qps,
        itemStyle: { color: '#409EFF' },
        areaStyle: { opacity: 0.1 },
      },
      {
        name: 'Avg RT (ms)',
        type: 'line',
        smooth: true,
        data: avgRT,
        itemStyle: { color: '#67C23A' },
      },
      {
        name: 'Success Rate (%)',
        type: 'line',
        smooth: true,
        yAxisIndex: 1,
        data: successRates,
        itemStyle: { color: '#E6A23C' },
      },
    ],
  }
})
</script>

<template>
  <v-chart
    :option="chartOption"
    :style="{ height: height, width: '100%' }"
    autoresize
  />
</template>
