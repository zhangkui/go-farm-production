# BUG-008 评测说明

本分支包含项目、BUG-008 模型修复代码和本题公开回归测试。

- Go：go1.26.1 windows/amd64
- MySQL：Docker 默认 127.0.0.1:3306，用户 root，未设置 TEST_MYSQL_DSN。
- 验证：go test ./scripts/verify -count=1 -run '^TestBug008_BusinessRegression$'

Docker 构建：bash build_benzhi_docker.sh

问题现象见 BUG_REPRO.md，验证记录见 BENZHI_VALIDATION/BUG-008.md。
