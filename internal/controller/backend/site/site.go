package site

import (
	"context"
	v2 "jh_app_service/api/backend/site/v1"
	"jh_app_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v2.UnimplementedSiteServer
}

func Register(s *grpcx.GrpcServer) {
	v2.RegisterSiteServer(s.Server, &Controller{})
}

// GetBasicSetting 获取站点基本设置 (gRPC)
func (*Controller) GetBasicSetting(ctx context.Context, req *v2.GetBasicSettingReq) (res *v2.GetBasicSettingRes, err error) {
	return backend.Site().GetBasicSetting(ctx, req)
}

// UpdateBasicSetting 更新站点基本设置 (gRPC)
func (*Controller) UpdateBasicSetting(ctx context.Context, req *v2.UpdateBasicSettingReq) (res *v2.UpdateBasicSettingRes, err error) {
	return backend.Site().UpdateBasicSetting(ctx, req)
}

// GetInviteBonusConfig 获取邀请活动配置 (gRPC)
func (*Controller) GetInviteBonusConfig(ctx context.Context, req *v2.GetInviteBonusConfigReq) (res *v2.GetInviteBonusConfigRes, err error) {
	return backend.Site().GetInviteBonusConfig(ctx, req)
}
