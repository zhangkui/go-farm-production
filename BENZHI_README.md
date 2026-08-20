# BUG-001 评测说明

本分支包含农场生产管理 Go 项目、BUG-001 的模型修复代码和该题唯一公开验证测试。

## 环境

- Go：`go1.26.1 windows/amd64`
- MySQL：测试通过 `TEST_MYSQL_DSN` 连接可用的 MySQL 8 实例
- 固定测试入口：`go test ./scripts/verify -count=1 -run '^TestBug001_BusinessRegression$'`

## 本地验证

```bash
export TEST_MYSQL_DSN='root:rootsecret@tcp(127.0.0.1:3307)/?parseTime=true&multiStatements=true&charset=utf8mb4'
go test ./scripts/verify -count=1 -run '^TestBug001_BusinessRegression$'
```

## Docker 构建

```bash
bash build_benzhi_docker.sh
```

问题现象和复现边界见 `BUG_REPRO.md`，真实验证记录见 `BENZHI_VALIDATION/BUG-001.md`。
