package user

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/gogf/gf/v2/frame/g"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	v1 "jh_app_service/api/frontend/user/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/service/frontend"
	"jh_app_service/internal/tracing"
)

type sUser struct{}

func New() *sUser {
	return &sUser{}
}

func init() {
	frontend.RegisterUser(New())
}

// GetUserInfo 获取用户信息
func (s *sUser) GetUserInfo(ctx context.Context, req *v1.GetUserInfoReq) (res *v1.GetUserInfoRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.GetUserInfo", trace.WithAttributes(
		attribute.String("method", "GetUserInfo"),
	))
	defer span.End()

	// 从JWT token获取用户ID
	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		middleware.LogWithTrace(ctx, "error", "未获取用户ID")
		return nil, fmt.Errorf("未登录")
	}

	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	// 查询用户信息
	var user *entity.User
	err = dao.User.Ctx(ctx).Where("id", userId).
		Where("site_id", siteId).
		Scan(&user)

	if err != nil || user == nil {
		middleware.LogWithTrace(ctx, "error", "用户不存在: %d", userId)
		return nil, fmt.Errorf("用户不存在")
	}

	// 查询用户等级名称
	var userGrade *entity.UserGrade
	gradeName := ""
	if user.GradeId > 0 {
		err = dao.UserGrade.Ctx(ctx).Where("id", user.GradeId).
			Where("site_id", siteId).
			Scan(&userGrade)
		if err == nil && userGrade != nil {
			gradeName = userGrade.Name
		}
	}

	// 检查是否绑定安全问题
	bindSafeQuestion := 0
	if user.SafeAnswer != "" {
		bindSafeQuestion = 1
	}

	// 检查是否绑定资金密码
	bindPayPassword := 0
	if user.PayPassword != "" {
		bindPayPassword = 1
	}

	// 检查是否绑定银行卡
	bindBank := 0
	var bankCount int
	bankCount, err = dao.UserBank.Ctx(ctx).Where("user_id", userId).
		Where("site_id", siteId).
		Count()
	if err == nil && bankCount > 0 {
		bindBank = 1
	}

	// 构建响应
	res = &v1.GetUserInfoRes{
		Username:          user.Username,
		Realname:          user.Realname,
		Email:             user.Email,
		Mobile:            user.Mobile,
		Qq:                user.Qq,
		Sex:               int32(user.Sex),
		Birthday:          gconv.String(user.Birthday),
		LastLoginTime:     user.LastLoginTime.String(),
		LastLoginAddress:  user.LastLoginAddress,
		ShowBeginnerGuide: int32(user.ShowBeginnerGuide),

		// 余额相关（先写死默认值）
		Balance:       "0.00",
		BalanceFrozen: "0.00",
		Points:        "0.00",

		// 等级相关
		GradeId:   int32(user.GradeId),
		GradeName: gradeName,

		// 绑定状态
		BindSafeQuestion: int32(bindSafeQuestion),
		BindPayPassword:  int32(bindPayPassword),
		BindBank:         int32(bindBank),

		// 其他（先写死默认值）
		UnreadMessageCount:  0,
		SwitchSign:          0,
		ApplyBirthdayBonus:  0,
		AutoBindGameBalance: 0,

		// 统计数据（先写死默认值）
		TotalBet:    "0.00",
		TotalBonus:  "0.00",
		TotalRebate: "0.00",

		// 最后记录（先写死空数组）
		LastRecharge: []*v1.LastRecord{},
		LastWithdraw: []*v1.LastRecord{},
		LastMessage:  []*v1.LastMessage{},
	}

	g.Log().Infof(ctx, "获取用户信息成功: user_id=%d, username=%s", userId, user.Username)

	return res, nil
}

// Logout 用户退出登录
func (s *sUser) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.Logout", trace.WithAttributes(
		attribute.String("method", "Logout"),
	))
	defer span.End()

	// 从JWT token获取用户ID（可选，用于记录日志）
	userId, exists := middleware.GetUserIdFromContext(ctx)
	if exists {
		middleware.LogWithTrace(ctx, "info", "用户退出登录: user_id=%d", userId)
	}

	// JWT模式下，退出登录主要由前端删除token完成
	// 后端可以在这里做一些清理工作，比如：
	// 1. 记录退出日志
	// 2. 清理在线状态
	// 3. 通知其他服务（如Workerman）

	res = &v1.LogoutRes{
		Message: "退出成功",
	}

	return res, nil
}
