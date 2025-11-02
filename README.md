# kratos-cli-boost
kratos cli 增强版

# 安装
```
go install github.com/enneket/kratos-cli-boost@latest
```

# 使用
```
# 生成 proto 模板
kratos proto add api/helloworld/helloworld.proto
# 生成 proto 源码
kratos proto client api/helloworld/helloworld.proto
# 生成 server 模板
kratos proto server api/helloworld/helloworld.proto -t internal/service
# 生成 biz 模板
kratos proto biz api/helloworld/helloworld.proto -t internal/biz
# 生成 data 模板
kratos proto data api/helloworld/helloworld.proto -t internal/data
```