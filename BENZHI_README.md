# BUG-005 评测说明

本分支包含项目、本题公开回归测试以及统一评测交付文件。Diagnosis 模型会话没有修改 Go 生产代码或测试文件。

- Go：go1.26.1 windows/amd64
- MySQL：Docker 默认 127.0.0.1:3306，用户 root，未设置 TEST_MYSQL_DSN。
- 验证：go test ./scripts/verify -count=1 -run '^TestBug005_BusinessRegression$'

Docker 构建：bash build_benzhi_docker.sh

问题现象见 BUG_REPRO.md，诊断验证记录见 BENZHI_VALIDATION/BUG-005.md。
