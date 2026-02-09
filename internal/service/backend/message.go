// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ================================================================================

package backend

import (
	"context"
	"jh_app_service/api/backend/message/v1"
)

type (
	IMessage interface {
		GetMessageList(ctx context.Context, req *v1.GetMessageListReq) (*v1.GetMessageListRes, error)
		CreateMessage(ctx context.Context, req *v1.CreateMessageReq) (*v1.CreateMessageRes, error)
		UpdateMessage(ctx context.Context, req *v1.UpdateMessageReq) (*v1.UpdateMessageRes, error)
		DeleteMessage(ctx context.Context, req *v1.DeleteMessageReq) (*v1.DeleteMessageRes, error)
		GetUserMessages(ctx context.Context, req *v1.GetUserMessagesReq) (*v1.GetUserMessagesRes, error)
		GetUnreadCount(ctx context.Context, req *v1.GetUnreadCountReq) (*v1.GetUnreadCountRes, error)
		ReadMessage(ctx context.Context, req *v1.ReadMessageReq) (*v1.ReadMessageRes, error)
	}
)

var (
	localMessage IMessage
)

func Message() IMessage {
	if localMessage == nil {
		panic("implement not found for interface IMessage, forgot register?")
	}
	return localMessage
}

func RegisterMessage(i IMessage) {
	localMessage = i
}
