# 农场种植与投入品管理系统 (go-farm-production)

面向规模化农场的全流程生产管理与成本核算系统：地块管理 → 种植计划 → 农事任务 → 投入品库存 → 采收入库 → 成本分析，全链路可追溯。

- **后端**：Go 1.22+ · 标准库 `net/http` + gorilla/mux · 分层架构 (transport → service → repository)
- **前端**：Vue 3 · TypeScript · Vite · Element Plus · Pinia · ECharts
- **存储**：MySQL 8.0.46 · Redis 7.4.10
- **部署**：Docker Compose 全容器化

## 目录结构

```
go-farm-production/
├── cmd/server/            # 程序入口
├── internal/
│   ├── app/               # 依赖装配 (DB/Redis/migrate/seed/HTTP)
│   ├── config/            # 配置加载 (viper, 环境变量优先)
│   ├── domain/            # 实体、枚举、错误、分页、Decimal
│   ├── migrate/           # 嵌入式版本化 SQL 迁移
│   ├── repository/        # 数据访问层 (MySQL + Redis 缓存)
│   ├── service/           # 业务逻辑 + 事务 + 审计
│   └── transport/http/    # 路由、中间件、handler、统一响应
├── migrations/            # 可直接执行的 SQL 迁移脚本 (与内嵌副本同步)
├── web/                   # Vue 3 前端
├── configs/config.yaml    # 本地开发配置
├── docker-compose.yml
├── Dockerfile
└── .env.example
```

## 快速开始

### 方式一：Docker Compose（推荐）

```bash
cp .env.example .env       # 按需修改默认管理员密码 / JWT 密钥
docker compose up -d --build
```

服务启动后：

- 前端：<http://localhost>
- API：<http://localhost:8080>
- 健康检查：<http://localhost:8080/health>
- 默认管理员：`admin` / `Admin123!`（首次启动自动初始化）

停止：`docker compose down`；重置数据：`docker compose down -v`。

### 方式二：本地开发

后端：

```bash
go mod download
go run ./cmd/server           # 依赖 configs/config.yaml 或环境变量
```

前端：

```bash
cd web
npm install
npm run dev                  # http://localhost:5173 (代理 /api -> :8080)
```

## 核心业务规则

| 不变量 | 实现位置 |
|--------|----------|
| 同地块种植计划时间不重叠 | `service/planting_plan.go` + `repository.FindOverlap` (半开区间) |
| 投入品领用与库存扣减事务化 | `service/allocation.go` 单事务 + 行级 `remaining_qty >= qty` 守卫 |
| 投入品剩余非负 | `repository/batch.go` `DecreaseStock` 守卫 UPDATE |
| 状态合法流转 | `domain/enums.go` `AllowedPlanTransition` / `service` 状态机 |
| 采收明细重量之和 = 总重量 | `service/harvest.go` `validateHarvest` |
| 成本自动计算 (人工+投入+设备+其他) | `service/cost_analysis.go` 聚合 |
| 外键参照完整性 | 迁移脚本 `ON DELETE RESTRICT` |

## 角色与权限

内置 6 个角色：`admin` / `farm_manager` / `production_supervisor` / `operator` / `auditor` / `read_only`，权限以 `resource:action` 表达式授予（详见 `migrations/0002_seed.sql`）。后端 RBAC 中间件在 Service 层强制校验，前端 `v-permission` 指令仅用于隐藏 UI。

## API 规范

- 版本前缀 `/api/v1/`，资源复数名词
- 统一响应信封：`{ code, message, data, request_id }`
- 分页：`page` / `page_size`，排序字段白名单校验防注入
- 写操作可携带 `Idempotency-Key` 头防重复提交
- JWT：Access Token 15 分钟，Refresh Token 7 天，刷新即轮换
- 登录限流：同 IP 5 次/分钟，同用户 10 次/分钟

## 质量

```bash
go vet ./...                # 静态检查
gofmt -l .                  # 格式化检查
go test ./...               # 单元测试
cd web && npm run type-check && npm run build
```

## 交付物清单

- ✅ Go 后端源码（≥51 文件 / ≥5000 行）
- ✅ Vue 3 前端源码
- ✅ 版本化数据库迁移脚本
- ✅ Docker Compose 一键部署
- ✅ .env.example 环境变量示例
- ✅ 默认管理员自动初始化

## 生产部署建议

- 强制 HTTPS，配置 SSL 证书
- 修改默认管理员密码，使用强随机 JWT 密钥 (≥32 字节)
- 配置 MySQL 定期备份与主从复制
- 监控系统资源与日志聚合 (ELK / Loki)
- 生产环境设置 `APP_ENV=prod`，`LOG_FORMAT=json`
