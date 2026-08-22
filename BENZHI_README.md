# Evaluation Guide

- Repository: zhangkui/go-farm-production
- Branch: test_model_fix7

## Public verification

```bash
go test ./scripts/verify -count=1 -run '^TestBug007_BusinessRegression$'
```
