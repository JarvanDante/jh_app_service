package public

import (
	"context"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"

	v1 "jh_app_service/api/frontend/public/v1"
	"jh_app_service/internal/service/frontend"
)

type Controller struct {
	v1.UnimplementedPublicServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterPublicServer(s.Server, &Controller{})
}

// QuickRegister 快速注册
func (c *Controller) QuickRegister(ctx context.Context, req *v1.QuickRegisterReq) (*v1.QuickRegisterRes, error) {
	return frontend.Public().QuickRegister(ctx, req)
}
