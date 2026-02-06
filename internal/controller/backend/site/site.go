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

// GetSiteActivityList 查询活动配置列表 (gRPC)
func (*Controller) GetSiteActivityList(ctx context.Context, req *v2.GetSiteActivityListReq) (res *v2.GetSiteActivityListRes, err error) {
	return backend.Site().GetSiteActivityList(ctx, req)
}

// UpdateSiteActivity 更新活动配置 (gRPC)
func (*Controller) UpdateSiteActivity(ctx context.Context, req *v2.UpdateSiteActivityReq) (res *v2.UpdateSiteActivityRes, err error) {
	return backend.Site().UpdateSiteActivity(ctx, req)
}
