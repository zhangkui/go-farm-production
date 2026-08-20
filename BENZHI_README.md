# BUG-002 评测说明

本分支包含 go-farm-production 项目、BUG-002 的模型修复代码和本题公开回归测试。

## 环境

- Go：`go1.26.1 windows/amd64`
- MySQL：测试使用 Docker 默认 MySQL，地址 `127.0.0.1:3306`，用户 `root`；本次验证未设置 `TEST_MYSQL_DSN`。
- 固定验证入口：`go test ./scripts/verify -count=1 -run '^TestBug002_BusinessRegression$'`

## 本地验证

```bash
go test ./scripts/verify -count=1 -run '^TestBug002_BusinessRegression$'
```

## Docker 构建

```bash
bash build_benzhi_docker.sh
```

问题现象、触发步骤和正确结果见 `BUG_REPRO.md`；真实验证记录见 `BENZHI_VALIDATION/BUG-002.md`。
