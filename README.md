# JH App Service - 前台用户/后台管理员微服务(private)

基于 GoFrame 框架的前台用户/后台管理员微服务，提供 gRPC 接口。

## 功能特性

- **gRPC 服务** - 提供高性能的 gRPC 接口
- **用户管理** - 用户的增删改查操作
- **服务注册** - 自动注册到 Consul 服务发现
- **数据库支持** - 集成 MySQL 数据库
- **配置管理** - 支持 YAML 配置文件

## 项目结构

```
jh_app_service
│ ├── Makefile
│ ├── README.md
│ ├── api
│ │ └── backend
│ ├── go.mod
│ ├── go.sum
│ ├── hack
│ │ └── config.yaml
│ ├── internal
│ │ ├── cmd
│ │ ├── consts
│ │ ├── controller
│ │ ├── dao
│ │ ├── logic
│ │ ├── middleware
│ │ ├── model
│ │ ├── registry
│ │ ├── service
│ │ └── tracing
│ ├── jh_app_service
│ ├── jh_app_service_debug
│ ├── main.go
│ ├── manifest
│ │ ├── config
│ │ ├── deploy
│ │ ├── docker
│ │ ├── protobuf
│ │ └── sql
│ ├── scripts
│ │ ├── fix_pb_omitempty.sh
│ │ └── gen_pb_clean.sh
│ └── test
│     └── test_test.go
```

## 配置说明

### manifest/config/config.yaml

```yaml
# GRPC Server
grpc:
  address: ":8002" # gRPC 服务端口
  name: "jh_app_service" # 服务名称
  logPath: "" # 日志路径
  logStdout: true # 控制台输出
  errorStack: true # 错误堆栈
  errorLogEnabled: true # 错误日志
  errorLogPattern: "error-{Ymd}.log"
  accessLogEnabled: true # 访问日志
  accessLogPattern: "access-{Ymd}.log"

# Consul 配置
consul:
  addr: "172.19.0.29:8500" # Consul 地址
  serviceName: "admin_service" # 在 Consul 中的服务名
  servicePort: 8002 # 服务端口
  serviceAddr: "172.19.0.31" # 服务地址

# 日志配置
logger:
  level: "all" # 日志级别
  stdout: true # 控制台输出

# 数据库配置
database:
  logger:
    level: "all"
    stdout: true
  default:
    link: "mysql:root:654321@tcp(172.19.0.4:3306)/test"
    debug: true
```

## gRPC 接口

### 用户服务 (User Service)

#### 方法列表

1. **Create** - 创建用户

   ```protobuf
   rpc Create(CreateReq) returns (CreateRes)
   ```

2. **GetOne** - 获取单个用户

   ```protobuf
   rpc GetOne(GetOneReq) returns (GetOneRes)
   ```

3. **GetList** - 获取用户列表

   ```protobuf
   rpc GetList(GetListReq) returns (GetListRes)
   ```

4. **Delete** - 删除用户
   ```protobuf
   rpc Delete(DeleteReq) returns (DeleteRes)
   ```

#### 消息定义

```protobuf
// 创建用户请求
message CreateReq {
    string Passport = 1; // 用户名
    string Password = 2; // 密码
    string Nickname = 3; // 昵称
}

// 获取用户请求
message GetOneReq {
    uint64 Id = 1; // 用户ID
}

// 获取用户列表请求
message GetListReq {
    int32 Page = 1; // 页码
    int32 Size = 2; // 每页大小
}

// 删除用户请求
message DeleteReq {
    uint64 Id = 1; // 用户ID
}
```

## 部署运行

### 1. Docker 方式（推荐）

```bash
# 在根目录启动所有服务
docker-compose up --build

# 单独启动用户服务容器
docker-compose up jh_app_service

# 进入容器
docker-compose exec jh_app_service bash

# 在容器内启动服务
cd /var/www/html/jh/jh_app_service
go run main.go
```

### 2. 本地开发

```bash
# 安装依赖
go mod tidy

# 生成 Proto 代码（如果需要）
protoc --go_out=. --go-grpc_out=. manifest/protobuf/user/v1/user.proto

# 运行服务
go run main.go

# 调试模式
dlv debug --headless --listen=:2345 --api-version=2
```

## 数据库配置

### 1. 创建数据库

```sql
CREATE DATABASE test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 创建用户表

```sql
CREATE TABLE `user` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `passport` varchar(45) NOT NULL COMMENT '用户账号',
  `password` varchar(45) NOT NULL COMMENT '用户密码',
  `nickname` varchar(45) NOT NULL COMMENT '用户昵称',
  `create_at` datetime DEFAULT NULL COMMENT '创建时间',
  `update_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `passport` (`passport`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 服务注册

服务启动时会自动注册到 Consul：

1. **服务名**: `admin_service`
2. **地址**: `172.19.0.31:8002`
3. **健康检查**: TCP 连接检查
4. **检查间隔**: 15 秒
5. **超时时间**: 10 秒

## gRPC 客户端测试

### 使用 GoLand gRPC 客户端

1. 配置 Proto 文件路径：

   ```
   /path/to/jh_app_service/manifest/protobuf/user/v1/user.proto
   ```

2. 设置依赖目录：

   ```
   /path/to/jh_app_service/manifest/protobuf
   ```

3. 连接地址：
   ```
   172.19.0.31:8002
   ```

### 使用 grpcurl 测试

```bash
# 列出服务
grpcurl -plaintext 172.19.0.31:8002 list

# 调用 GetList 方法
grpcurl -plaintext -d '{"Page":1,"Size":10}' 172.19.0.31:8002 user.User/GetList

# 调用 GetOne 方法
grpcurl -plaintext -d '{"Id":1}' 172.19.0.31:8002 user.User/GetOne
```

## 开发指南

### 添加新的 gRPC 方法

1. 在 `manifest/protobuf/user/v1/user.proto` 中定义新方法
2. 重新生成 gRPC 代码
3. 在 `internal/controller/user/user.go` 中实现方法
4. 在 `internal/service/user.go` 中添加业务逻辑

### 修改数据库模型

1. 更新 `manifest/sql/create.sql`
2. 修改 `internal/model/` 中的模型定义
3. 更新 DAO 层的数据访问方法

## 监控与调试

### 服务状态检查

```bash
# 检查服务是否在 Consul 中注册
curl http://localhost:8500/v1/catalog/service/admin_service

# 检查服务健康状态
curl http://localhost:8500/v1/health/service/admin_service
```

### 日志查看

```bash
# Docker 方式
docker-compose logs -f jh_app_service

# 本地方式
查看控制台输出和日志文件
```

## 故障排查

### 常见问题

1. **服务启动失败**

   - 检查端口 8002 是否被占用
   - 确认数据库连接配置正确

2. **Consul 注册失败**

   - 检查 Consul 服务是否启动
   - 确认网络连接正常

3. **gRPC 连接失败**

   - 检查防火墙设置
   - 确认服务地址和端口正确

4. **数据库连接失败**
   - 检查 MySQL 服务状态
   - 确认数据库用户权限

## Proto 文件管理

### 生成 gRPC 代码

```bash
# 安装 protoc 编译器和插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成代码
protoc --go_out=. --go-grpc_out=. manifest/protobuf/user/v1/user.proto
protoc --go_out=. --go-grpc_out=. manifest/protobuf/pbentity/user.proto
```

### Proto 文件版本管理

- 使用语义化版本号 (v1, v2, ...)
- 保持向后兼容性
- 新增字段使用新的字段号

# 转盘配置
8 扇区示例（偏稳健）
扇区1：谢谢参与，0，权重35，上限0，VIP0，图标🙈
扇区2：₱2，2，权重25，上限0，VIP0，图标💵
扇区3：₱5，5，权重18，上限0，VIP0，图标🎁
扇区4：₱8，8，权重10，上限0，VIP0，图标💎
扇区5：₱10，10，权重6，上限0，VIP1，图标💰
扇区6：₱20，20，权重3，上限2，VIP2，图标🏆
扇区7：₱50，50，权重2，上限1，VIP4，图标👑
扇区8：₱88，88，权重1，上限1，VIP6，图标🚀