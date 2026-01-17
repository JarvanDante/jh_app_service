# Balance Client Proto 说明

## 概述

为了避免服务间的模块依赖，我们将 `jh_balance_service` 的 protobuf 定义复制到 `jh_app_service` 中作为客户端使用。

## 目录结构

```
jh_app_service/
├── manifest/protobuf/frontend/balance/v1/
│   └── balance.proto                    # 从 balance_service 复制的 proto 文件
└── api/frontend/balance/v1/
    ├── balance.pb.go                    # 生成的 Go 代码
    └── balance_grpc.pb.go               # 生成的 gRPC 客户端代码
```

## 使用方式

### 1. 导入包

```go
import balanceV1 "jh_app_service/api/frontend/balance/v1"
```

### 2. 创建 gRPC 客户端

```go
conn, err := grpc.NewClient(
    "jh_balance_service:8003",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
client := balanceV1.NewBalanceClient(conn)
```

### 3. 调用接口

```go
req := &balanceV1.InitBalanceUserReq{
    SiteId:   1,
    UserId:   123,
    Username: "test",
    Currency: "USD",
}
resp, err := client.InitBalanceUser(ctx, req)
```

## 更新流程

当 `jh_balance_service` 的 proto 文件发生变化时：

### 步骤 1：复制 proto 文件

```bash
cd jh_app_service
cp ../jh_balance_service/manifest/protobuf/frontend/balance/v1/balance.proto \
   manifest/protobuf/frontend/balance/v1/
```

### 步骤 2：修改 go_package 路径

编辑 `manifest/protobuf/frontend/balance/v1/balance.proto`：

```protobuf
package frontend_balance;  // 保持不变
option go_package = "jh_app_service/api/frontend/balance/v1";  // 改为 app_service 的路径
```

### 步骤 3：重新生成代码

```bash
./scripts/gen_pb_clean.sh
```

### 步骤 4：验证编译

```bash
go build -o /dev/null ./internal/mq/...
```

## 优点

✅ **独立部署** - 各服务可以独立编译和部署
✅ **无模块依赖** - 不需要在 go.mod 中添加其他服务的依赖
✅ **版本控制** - 每个服务控制自己使用的 proto 版本
✅ **编译速度** - 不需要加载其他服务的代码
✅ **目录统一** - 使用标准的 frontend/balance 目录结构

## 注意事项

⚠️ **手动同步** - proto 文件变更需要手动复制和更新
⚠️ **版本一致性** - 需要确保客户端和服务端的 proto 定义一致
⚠️ **只改路径** - 复制后只需修改 `go_package`，package 名保持 `frontend_balance`

## 当前使用的接口

- `InitBalanceUser` - 初始化用户余额（用于快速注册后异步创建余额记录）

## 相关文件

- 源文件：`jh_balance_service/manifest/protobuf/frontend/balance/v1/balance.proto`
- 客户端文件：`jh_app_service/manifest/protobuf/frontend/balance/v1/balance.proto`
- 生成代码：`jh_app_service/api/frontend/balance/v1/`
- 使用位置：`jh_app_service/internal/mq/consumer.go`
