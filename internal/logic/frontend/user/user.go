package user

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"golang.org/x/crypto/bcrypt"

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

// BindEmail 绑定邮箱
func (s *sUser) BindEmail(ctx context.Context, req *v1.BindEmailReq) (res *v1.BindEmailRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.BindEmail", trace.WithAttributes(
		attribute.String("method", "BindEmail"),
	))
	defer span.End()

	// 验证邮箱格式
	if req.Email == "" {
		return nil, fmt.Errorf("邮箱必填")
	}

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

	// 检查邮箱是否已被其他用户绑定
	var existUser *entity.User
	err = dao.User.Ctx(ctx).Where("email", req.Email).
		Where("site_id", siteId).
		Where("id !=", userId).
		Scan(&existUser)

	if err == nil && existUser != nil {
		return nil, fmt.Errorf("该邮箱已被其他用户绑定")
	}

	// 更新用户邮箱
	_, err = dao.User.Ctx(ctx).Where("id", userId).
		Where("site_id", siteId).
		Update(map[string]interface{}{
			"email": req.Email,
		})

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "绑定邮箱失败: %v", err)
		return nil, fmt.Errorf("绑定邮箱失败")
	}

	middleware.LogWithTrace(ctx, "info", "绑定邮箱成功: user_id=%d, email=%s", userId, req.Email)

	res = &v1.BindEmailRes{
		Message: "绑定邮箱成功",
	}

	return res, nil
}

// BindMobile 绑定手机
func (s *sUser) BindMobile(ctx context.Context, req *v1.BindMobileReq) (res *v1.BindMobileRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.BindMobile", trace.WithAttributes(
		attribute.String("method", "BindMobile"),
	))
	defer span.End()

	// 验证手机号
	if req.Mobile == "" {
		return nil, fmt.Errorf("手机号必填")
	}

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

	// 检查手机号是否已被其他用户绑定
	var existUser *entity.User
	err = dao.User.Ctx(ctx).Where("mobile", req.Mobile).
		Where("site_id", siteId).
		Where("id !=", userId).
		Scan(&existUser)

	if err == nil && existUser != nil {
		return nil, fmt.Errorf("该手机号已被其他用户绑定")
	}

	// 更新用户手机号
	_, err = dao.User.Ctx(ctx).Where("id", userId).
		Where("site_id", siteId).
		Update(map[string]interface{}{
			"mobile": req.Mobile,
		})

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "绑定手机号失败: %v", err)
		return nil, fmt.Errorf("绑定手机号失败")
	}

	middleware.LogWithTrace(ctx, "info", "绑定手机号成功: user_id=%d, mobile=%s", userId, req.Mobile)

	res = &v1.BindMobileRes{
		Message: "绑定手机号成功",
	}

	return res, nil
}

// ChangeLoginPassword 修改登录密码
func (s *sUser) ChangeLoginPassword(ctx context.Context, req *v1.ChangeLoginPasswordReq) (res *v1.ChangeLoginPasswordRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.ChangeLoginPassword", trace.WithAttributes(
		attribute.String("method", "ChangeLoginPassword"),
	))
	defer span.End()

	// 验证参数
	if req.OldPassword == "" {
		return nil, fmt.Errorf("请输入原密码")
	}

	if req.NewPassword == "" {
		return nil, fmt.Errorf("请输入新密码")
	}

	if len(req.NewPassword) < 6 {
		return nil, fmt.Errorf("新密码最少需要6个字符长度")
	}

	if len(req.NewPassword) > 20 {
		return nil, fmt.Errorf("新密码最多不超过20个字符长度")
	}

	if req.OldPassword == req.NewPassword {
		return nil, fmt.Errorf("新密码与原密码必须不一致")
	}

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

	// 验证旧密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "原密码错误: user_id=%d", userId)
		return nil, fmt.Errorf("原密码错误")
	}

	// 使用bcrypt加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "密码加密失败: %v", err)
		return nil, fmt.Errorf("修改登录密码失败")
	}

	// 更新密码
	_, err = dao.User.Ctx(ctx).Where("id", userId).
		Where("site_id", siteId).
		Update(map[string]interface{}{
			"password": string(hashedPassword),
		})

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "修改登录密码失败: %v", err)
		return nil, fmt.Errorf("修改登录密码失败")
	}

	middleware.LogWithTrace(ctx, "info", "修改登录密码成功: user_id=%d", userId)

	res = &v1.ChangeLoginPasswordRes{
		Message: "修改登录密码成功",
	}

	return res, nil
}

// BindLoginPassword 绑定登录密码（针对快速注册用户）
func (s *sUser) BindLoginPassword(ctx context.Context, req *v1.BindLoginPasswordReq) (res *v1.BindLoginPasswordRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.BindLoginPassword", trace.WithAttributes(
		attribute.String("method", "BindLoginPassword"),
	))
	defer span.End()

	// 验证参数
	if req.NewPassword == "" {
		return nil, fmt.Errorf("请输入新密码")
	}

	if len(req.NewPassword) < 6 {
		return nil, fmt.Errorf("新密码最少需要6个字符长度")
	}

	if len(req.NewPassword) > 20 {
		return nil, fmt.Errorf("新密码最多不超过20个字符长度")
	}

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

	if user.HasPass > 0 {
		middleware.LogWithTrace(ctx, "error", "已经绑定过密码: user_id=%d", userId)
		return nil, fmt.Errorf("已经绑定过密码请勿重新修改")
	}

	// 使用bcrypt加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "密码加密失败: %v", err)
		return nil, fmt.Errorf("绑定登录密码失败")
	}

	// 更新密码
	_, err = dao.User.Ctx(ctx).Where("id", userId).
		Where("site_id", siteId).
		Update(map[string]interface{}{
			"password": string(hashedPassword),
			"has_pass": 1,
		})

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "绑定登录密码失败: %v", err)
		return nil, fmt.Errorf("绑定登录密码失败")
	}

	middleware.LogWithTrace(ctx, "info", "绑定登录密码成功: user_id=%d", userId)

	res = &v1.BindLoginPasswordRes{
		Message: "绑定登录密码成功",
	}

	return res, nil
}

// GetGradeList 获取会员等级列表
func (s *sUser) GetGradeList(ctx context.Context, req *v1.GetGradeListReq) (res *v1.GetGradeListRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.GetGradeList", trace.WithAttributes(
		attribute.String("method", "GetGradeList"),
	))
	defer span.End()

	// 从JWT token获取站点ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	// 查询会员等级列表
	var grades []*entity.UserGrade
	err = dao.UserGrade.Ctx(ctx).
		Where("site_id", siteId).
		Where("status", 1). // STATUS_ACTIVE = 1
		OrderAsc("id").
		Scan(&grades)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询会员等级列表失败: %v", err)
		return nil, fmt.Errorf("查询会员等级列表失败")
	}

	// 构建响应
	res = &v1.GetGradeListRes{
		List: make([]*v1.GradeItem, 0, len(grades)),
	}

	for _, grade := range grades {
		item := &v1.GradeItem{
			Id:                   int32(grade.Id),
			Name:                 grade.Name,
			DepositAmount:        fmt.Sprintf("%.2f", grade.DepositAmount), // 充值金额
			BetAmount:            fmt.Sprintf("%.2f", grade.BetAmount),     // 投注金额
			PointsUpgrade:        int32(grade.PointsUpgrade),
			BonusUpgrade:         fmt.Sprintf("%.2f", grade.BonusUpgrade),
			BonusBirthday:        fmt.Sprintf("%.2f", grade.BonusBirthday),
			RebatePercentSports:  fmt.Sprintf("%.2f", grade.RebatePercentSports*100),
			RebatePercentLottery: fmt.Sprintf("%.2f", grade.RebatePercentLottery*100),
			RebatePercentLive:    fmt.Sprintf("%.2f", grade.RebatePercentLive*100),
			RebatePercentEgame:   fmt.Sprintf("%.2f", grade.RebatePercentEgame*100),
			RebatePercentPoker:   fmt.Sprintf("%.2f", grade.RebatePercentPoker*100),
		}
		res.List = append(res.List, item)
	}

	middleware.LogWithTrace(ctx, "info", "获取会员等级列表成功: site_id=%d, count=%d", siteId, len(grades))

	return res, nil
}
