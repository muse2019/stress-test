# Stress Test Tool

HTTP 接口压测工具，支持 Web 控制台可视化操作，实时查看压测结果。

## 功能特性

- **固定并发压测**：指定并发数和持续时间进行压测
- **实时监控**：通过 WebSocket 实时推送 QPS、响应时间、成功率等指标
- **百分位统计**：支持 P50、P90、P95、P99 响应时间统计
- **历史报告**：自动生成 Markdown 格式的压测报告
- **任务管理**：支持任务的创建、编辑、删除、启动、停止

## 技术栈

### 后端
- Go 1.21+
- gorilla/mux (HTTP 路由)
- gorilla/websocket (WebSocket)
- Ring Buffer 统计收集器

### 前端
- Vue 3 + TypeScript
- Element Plus UI
- ECharts 图表
- Vite 构建工具

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+

### 启动后端

```bash
cd stress-test
go mod download
go run ./cmd/server
```

后端服务将在 `http://localhost:8080` 启动。

### 启动前端

```bash
cd stress-test/web
npm install
npm run dev
```

前端服务将在 `http://localhost:3344` 启动。

## 使用方法

1. 打开浏览器访问 `http://localhost:3344`
2. 在"任务管理"页面创建新的压测任务
3. 点击"启动"开始压测
4. 在"实时监控"页面查看压测进度和实时指标
5. 压测完成后，在"历史报告"页面查看详细报告

## API 接口

### 任务管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/tasks | 获取任务列表 |
| POST | /api/tasks | 创建任务 |
| GET | /api/tasks/{id} | 获取任务详情 |
| PUT | /api/tasks/{id} | 更新任务 |
| DELETE | /api/tasks/{id} | 删除任务 |
| POST | /api/tasks/{id}/start | 启动压测 |
| POST | /api/tasks/{id}/stop | 停止压测 |

### 报告管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/reports | 获取报告列表 |
| GET | /api/reports/{id} | 获取报告详情 |
| GET | /api/reports/{id}/download | 下载 Markdown 报告 |

### WebSocket

连接地址：`ws://localhost:8080/ws/tasks/{id}`

消息类型：
- `started` - 压测开始
- `stats` - 实时统计数据（每秒推送）
- `completed` - 压测完成
- `error` - 错误信息

## 数据存储

- 任务配置：`data/tasks.json`
- 压测报告：`data/reports/{report-id}.json` 和 `data/reports/{report-id}.md`

## 项目结构

```
stress-test/
├── cmd/
│   └── server/          # 服务入口
├── internal/
│   ├── api/             # HTTP API 和 WebSocket
│   ├── engine/          # HTTP 压测引擎
│   ├── scheduler/       # 调度器
│   ├── stats/           # 统计收集器
│   ├── store/           # 数据存储
│   └── report/          # 报告生成
├── pkg/
│   └── models/          # 数据模型
├── web/                 # 前端代码
│   ├── src/
│   │   ├── api/         # API 客户端
│   │   ├── components/  # 组件
│   │   ├── views/       # 页面
│   │   └── types/       # TypeScript 类型
│   └── ...
└── data/                # 数据目录
    ├── tasks.json
    └── reports/
```

## License

MIT
