package message

import (
	"context"
	v1 "jh_app_service/api/backend/message/v1"
	"jh_app_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v1.UnimplementedMessageServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterMessageServer(s.Server, &Controller{})
}

// GetMessageList 获取消息列表
func (*Controller) GetMessageList(ctx context.Context, req *v1.GetMessageListReq) (res *v1.GetMessageListRes, err error) {
	return backend.Message().GetMessageList(ctx, req)
}

// CreateMessage 创建消息
func (*Controller) CreateMessage(ctx context.Context, req *v1.CreateMessageReq) (res *v1.CreateMessageRes, err error) {
	return backend.Message().CreateMessage(ctx, req)
}

// UpdateMessage 更新消息
func (*Controller) UpdateMessage(ctx context.Context, req *v1.UpdateMessageReq) (res *v1.UpdateMessageRes, err error) {
	return backend.Message().UpdateMessage(ctx, req)
}

// DeleteMessage 删除消息
func (*Controller) DeleteMessage(ctx context.Context, req *v1.DeleteMessageReq) (res *v1.DeleteMessageRes, err error) {
	return backend.Message().DeleteMessage(ctx, req)
}

// GetUserMessages 获取用户消息
func (*Controller) GetUserMessages(ctx context.Context, req *v1.GetUserMessagesReq) (res *v1.GetUserMessagesRes, err error) {
	return backend.Message().GetUserMessages(ctx, req)
}

// GetUnreadCount 获取未读数量
func (*Controller) GetUnreadCount(ctx context.Context, req *v1.GetUnreadCountReq) (res *v1.GetUnreadCountRes, err error) {
	return backend.Message().GetUnreadCount(ctx, req)
}

// ReadMessage 读取消息
func (*Controller) ReadMessage(ctx context.Context, req *v1.ReadMessageReq) (res *v1.ReadMessageRes, err error) {
	return backend.Message().ReadMessage(ctx, req)
}
