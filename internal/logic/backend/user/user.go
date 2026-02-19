package user

import (
	"context"
	"fmt"
	"strings"
	v1 "jh_app_service/api/backend/user/v1"
	logic_mq_outbox "jh_app_service/internal/logic/mq_outbox"
	"jh_app_service/internal/registry"
	"jh_app_service/internal/service/backend"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"

	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/tracing"
)

type (
	sUser struct{}
)

func init() {
	backend.RegisterUser(&sUser{})
}

// GetUserList 获取用户列表
func (s *sUser) GetUserList(ctx context.Context, req *v1.GetUserListReq) (*v1.GetUserListRes, error) {
	// 参数验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "user.GetUserList", trace.WithAttributes(
		attribute.String("method", "GetUserList"),
		attribute.String("username", req.Username),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("点ID不存在")
	}
	siteCode, err := dao.Site.GetCodeById(ctx, siteId)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "未获取站点code")
		return nil, fmt.Errorf("获取站点CODE失败: %v", err)
	}

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("配置的站点代码: %s", siteCode))

	tracing.SetSpanAttributes(span, attribute.Int("site_id", int(siteId)))

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("开始查询用户列表，站点ID: %d, 页码: %d, 每页: %d", siteId, req.Page, req.Size))

	// 设置默认分页参数
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	// 构建查询条件
	query := dao.User.Ctx(ctx).Where("site_id = ?", siteId)

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("基础查询条件 - 站点ID: %d", siteId))

	// 添加筛选条件
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.GradeId > 0 {
		query = query.Where("grade_id = ?", req.GradeId)
	}
	if req.Username != "" {
		query = query.Where("username = ?", req.Username)
	}
	if req.Mobile != "" {
		query = query.Where("mobile = ?", req.Mobile)
	}

	// 代理用户名筛选
	if req.AgentUsername != "" {
		// 先查找代理ID
		var agent entity.Agent
		err := dao.Agent.Ctx(ctx).Where("site_id = ? AND username = ?", siteId, req.AgentUsername).Scan(&agent)
		if err != nil {
			return &v1.GetUserListRes{List: []*v1.UserInfo{}, Count: 0}, nil
		}
		query = query.Where("agent_id = ?", agent.Id)
	}

	// 域名筛选
	if req.Domain != "" {
		query = query.WhereLike("register_url", "%"+req.Domain+"%")
	}

	// 时间范围筛选
	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	// 数据库查询span - 获取总数
	ctx, countSpan := tracing.StartSpan(ctx, "db.query.user_count", trace.WithAttributes(
		attribute.String("db.operation", "count"),
		attribute.String("db.table", "user"),
	))
	total, err := query.Count()
	countSpan.End()
	if err != nil {
		tracing.SetSpanError(span, err)
		return nil, fmt.Errorf("获取用户总数失败: %v", err)
	}

	// 数据库查询span - 获取列表数据
	ctx, listSpan := tracing.StartSpan(ctx, "db.query.user_list", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "user"),
		attribute.Int("limit", int(size)),
		attribute.Int("offset", int((page-1)*size)),
	))

	// 排序处理
	orderBy := "created_at DESC"
	if req.SortField != "" {
		orderField := s.getOrderByField(req.SortField)
		orderRule := "ASC"
		if req.SortRule == 0 {
			orderRule = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", orderField, orderRule)
	}

	var users []entity.User
	err = query.Page(int(page), int(size)).
		Order(orderBy).
		Scan(&users)
	listSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		return nil, fmt.Errorf("获取用户列表失败: %v", err)
	}

	// 转换为响应格式
	var userList []*v1.UserInfo
	for _, user := range users {
		userInfo := &v1.UserInfo{
			Id:         int32(user.Id),
			Username:   user.Username,
			GradeId:    int32(user.GradeId),
			AgentId:    int32(user.AgentId),
			Status:     int32(user.Status),
			RegisterIp: user.RegisterIp,
			Mobile:     s.maskMobile(user.Mobile),
			Email:      user.Email,
			IsOnline:   int32(user.IsOnline),
			PayTimes:   int32(user.PayTimes),
		}

		// 时间格式化（统一返回 yyyy-mm-dd hh:mm:ss）
		if user.RegisterTime != nil {
			userInfo.RegisterTime = strings.ReplaceAll(user.RegisterTime.Format("2006-01-02 15:04:05"), "/", "-")
		}
		if user.LastLoginTime != nil {
			userInfo.LastLoginTime = strings.ReplaceAll(user.LastLoginTime.Format("2006-01-02 15:04:05"), "/", "-")
		}

		// 获取等级名称
		var grade entity.UserGrade
		dao.UserGrade.Ctx(ctx).
			Where(dao.UserGrade.Columns().SiteId, siteId).
			Where(dao.UserGrade.Columns().Id, user.GradeId).
			Scan(&grade)
		userInfo.GradeName = grade.Name

		// 获取代理用户名
		var agent entity.Agent
		dao.Agent.Ctx(ctx).Where(do.Agent{
			SiteId: siteId,
			Id:     uint(user.AgentId),
		}).Scan(&agent)
		userInfo.AgentUsername = agent.AgentName

		userList = append(userList, userInfo)
	}

	// 确保返回空数组而不是nil
	if userList == nil {
		userList = []*v1.UserInfo{}
	}

	// 获取统计数据
	totalUsers, _ := dao.User.Ctx(ctx).Where("site_id = ?", siteId).Count()
	totalChargeUsers, _ := dao.User.Ctx(ctx).Where("site_id = ? AND pay_times >= 1", siteId).Count()

	tracing.SetSpanAttributes(span,
		attribute.Int("total_count", total),
		attribute.Int("returned_count", len(userList)),
	)

	return &v1.GetUserListRes{
		List:             userList,
		Count:            int32(total),
		TotalUsers:       int32(totalUsers),
		TotalChargeUsers: int32(totalChargeUsers),
	}, nil
}

// UpdateUser 更新用户信息
func (s *sUser) UpdateUser(ctx context.Context, req *v1.UpdateUserReq) (*v1.UpdateUserRes, error) {
	// 参数验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.Id <= 0 {
		return &v1.UpdateUserRes{Success: false, Message: "用户ID不能为空"}, nil
	}

	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "user.UpdateUser", trace.WithAttributes(
		attribute.String("method", "UpdateUser"),
		attribute.Int("user_id", int(req.Id)),
	))
	defer span.End()

	// 获取站点ID
	siteId := 1 // 临时硬编码

	// 查找用户
	var user entity.User
	err := dao.User.Ctx(ctx).Where("site_id = ? AND id = ?", siteId, req.Id).Scan(&user)

	if err != nil || user.Id == 0 {
		return &v1.UpdateUserRes{Success: false, Message: "用户不存在"}, nil
	}

	// 构建更新数据
	updateData := do.User{
		GradeId:   int(req.GradeId),
		AgentId:   int(req.AgentId),
		Status:    int(req.Status),
		UpdatedAt: gtime.Now(),
	}

	// 可选字段更新
	if req.LoginPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.LoginPassword), bcrypt.DefaultCost)
		if err != nil {
			return &v1.UpdateUserRes{Success: false, Message: "密码加密失败"}, nil
		}
		updateData.Password = string(hashedPassword)
	}

	if req.PayPassword != "" {
		hashedPayPassword, err := bcrypt.GenerateFromPassword([]byte(req.PayPassword), bcrypt.DefaultCost)
		if err != nil {
			return &v1.UpdateUserRes{Success: false, Message: "资金密码加密失败"}, nil
		}
		updateData.PayPassword = string(hashedPayPassword)
	}

	if req.Mobile != "" {
		updateData.Mobile = req.Mobile
	}
	if req.Email != "" {
		updateData.Email = req.Email
	}

	if req.Birthday != "" {
		birthday, err := gtime.StrToTime(req.Birthday)
		if err == nil {
			updateData.Birthday = birthday
		}
	}
	if req.Remark != "" {
		updateData.Remark = req.Remark
	}

	// 执行更新
	_, err = dao.User.Ctx(ctx).Where(do.User{
		SiteId: siteId,
		Id:     uint(req.Id),
	}).Update(updateData)

	if err != nil {
		tracing.SetSpanError(span, err)
		return &v1.UpdateUserRes{Success: false, Message: "更新用户失败"}, nil
	}

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("更新用户成功，用户ID: %d", req.Id))

	return &v1.UpdateUserRes{Success: true, Message: "更新用户成功"}, nil
}

// GetUserBasicInfo 获取用户基本信息
func (s *sUser) GetUserBasicInfo(ctx context.Context, req *v1.GetUserBasicInfoReq) (*v1.GetUserBasicInfoRes, error) {
	// 参数验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.Id <= 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "user.GetUserBasicInfo", trace.WithAttributes(
		attribute.String("method", "GetUserBasicInfo"),
		attribute.Int("user_id", int(req.Id)),
	))
	defer span.End()

	// 从 context 获取站点ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("站点ID不存在")
	}

	// 获取站点代码（用于动态数据库切换）
	siteCode, err := dao.Site.GetCodeById(ctx, siteId)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "未获取站点code")
		return nil, fmt.Errorf("获取站点CODE失败: %v", err)
	}

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("配置的站点代码: %s, 站点ID: %d", siteCode, siteId))
	tracing.SetSpanAttributes(span, attribute.Int("site_id", int(siteId)))

	// 查找用户
	var user entity.User
	err = dao.User.Ctx(ctx).Where("site_id = ? AND id = ?", siteId, req.Id).Scan(&user)

	if err != nil || user.Id == 0 {
		middleware.LogWithTrace(ctx, "error", fmt.Sprintf("用户不存在: userId=%d, siteId=%d", req.Id, siteId))
		return nil, fmt.Errorf("用户不存在")
	}

	// 构建基本信息
	basicInfo := &v1.UserBasicInfo{
		Id:          int32(user.Id),
		Username:    user.Username,
		GradeId:     int32(user.GradeId), // 添加等级ID
		Mobile:      user.Mobile,
		Email:       user.Email,
		Status:      int32(user.Status),
		Remark:      user.Remark,
		RegisterIp:  user.RegisterIp,
		LastLoginIp: user.LastLoginIp,
		IsOnline:    int32(user.IsOnline),
	}

	// 时间格式化
	if user.RegisterTime != nil {
		basicInfo.RegisterTime = user.RegisterTime.Format("2006-01-02 15:04:05")
	}
	if user.LastLoginTime != nil {
		basicInfo.LastLoginTime = user.LastLoginTime.Format("2006-01-02 15:04:05")
	}
	if user.Birthday != nil {
		basicInfo.Birthday = user.Birthday.Format("2006-01-02")
	}

	// 获取等级名称（使用动态数据库）
	var grade entity.UserGrade
	dao.UserGrade.Ctx(ctx).
		Where(dao.UserGrade.Columns().SiteId, siteId).
		Where(dao.UserGrade.Columns().Id, user.GradeId).
		Scan(&grade)
	basicInfo.GradeName = grade.Name

	// 获取代理名称
	var agent entity.Agent
	dao.Agent.Ctx(ctx).Where(do.Agent{
		SiteId: siteId,
		Id:     uint(user.AgentId),
	}).Scan(&agent)
	basicInfo.AgentName = agent.AgentName

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("获取用户基本信息成功，用户ID: %d, 用户名: %s, VIP等级: %d", req.Id, user.Username, user.GradeId))

	return &v1.GetUserBasicInfoRes{User: basicInfo}, nil
}

// getOrderByField 获取排序字段
func (s *sUser) getOrderByField(sortField string) string {
	switch sortField {
	case "register_time":
		return "register_time"
	case "balance":
		return "balance"
	case "balance_frozen":
		return "balance_frozen"
	case "last_login_time":
		return "last_login_time"
	case "points":
		return "points"
	default:
		return "created_at"
	}
}

// maskMobile 手机号脱敏
func (s *sUser) maskMobile(mobile string) string {
	if len(mobile) < 8 {
		return mobile
	}
	return mobile[:3] + "****" + mobile[len(mobile)-4:]
}

// GetUserGrades 获取用户等级列表
func (s *sUser) GetUserGrades(ctx context.Context, req *v1.GetUserGradesReq) (*v1.GetUserGradesRes, error) {
	// 参数验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "user.GetUserGrades", trace.WithAttributes(
		attribute.String("method", "GetUserGrades"),
		attribute.Int("site_id", int(req.SiteId)),
	))
	defer span.End()

	// 获取用户等级列表
	grades, err := s.getUserGradeList(ctx, int(req.SiteId))
	if err != nil {
		tracing.SetSpanError(span, err)
		return &v1.GetUserGradesRes{
			Code:    500,
			Message: "获取数据失败: " + err.Error(),
		}, nil
	}

	// 转换数据格式
	var gradeInfos []*v1.UserGradeInfo
	for _, grade := range grades {
		gradeInfo := &v1.UserGradeInfo{
			Id:                   int32(grade.Id),
			Name:                 grade.Name,
			RebatePercentSports:  grade.RebatePercentSports,
			RebatePercentLottery: grade.RebatePercentLottery,
			RebatePercentLive:    grade.RebatePercentLive,
			RebatePercentEgame:   grade.RebatePercentEgame,
			RebatePercentPoker:   grade.RebatePercentPoker,
			PaymentLimit:         grade.PaymentLimit,
			BetLimit:             grade.BetLimit,
			MoneyLimit:           grade.MoneyLimit,
			Activities:           []string{}, // TODO: 实现活动关联逻辑
		}
		gradeInfos = append(gradeInfos, gradeInfo)
	}

	tracing.SetSpanAttributes(span, attribute.Bool("success", true))
	return &v1.GetUserGradesRes{
		Code:    200,
		Message: "获取数据成功",
		Data:    gradeInfos,
	}, nil
}

// SaveUserGrades 保存用户等级
func (s *sUser) SaveUserGrades(ctx context.Context, req *v1.SaveUserGradesReq) (*v1.SaveUserGradesRes, error) {
	// 参数验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "user.SaveUserGrades", trace.WithAttributes(
		attribute.String("method", "SaveUserGrades"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.Int("grade_count", len(req.Data)),
	))
	defer span.End()

	// 转换数据格式
	var grades []*entity.UserGrade
	for _, gradeInfo := range req.Data {
		grade := &entity.UserGrade{
			Id:                   uint(gradeInfo.Id),
			Name:                 gradeInfo.Name,
			RebatePercentSports:  gradeInfo.RebatePercentSports,
			RebatePercentLottery: gradeInfo.RebatePercentLottery,
			RebatePercentLive:    gradeInfo.RebatePercentLive,
			RebatePercentEgame:   gradeInfo.RebatePercentEgame,
			RebatePercentPoker:   gradeInfo.RebatePercentPoker,
			PaymentLimit:         gradeInfo.PaymentLimit,
			BetLimit:             gradeInfo.BetLimit,
			MoneyLimit:           gradeInfo.MoneyLimit,
		}
		grades = append(grades, grade)
	}

	// 保存用户等级
	err := s.saveUserGrades(ctx, int(req.SiteId), grades)
	if err != nil {
		tracing.SetSpanError(span, err)
		return &v1.SaveUserGradesRes{
			Code:    500,
			Message: "保存失败: " + err.Error(),
		}, nil
	}

	tracing.SetSpanAttributes(span, attribute.Bool("success", true))
	return &v1.SaveUserGradesRes{
		Code:    200,
		Message: "保存成功",
	}, nil
}

// DeleteUserGrades 删除用户等级
func (s *sUser) DeleteUserGrades(ctx context.Context, req *v1.DeleteUserGradesReq) (*v1.DeleteUserGradesRes, error) {
	// 参数验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "user.DeleteUserGrades", trace.WithAttributes(
		attribute.String("method", "DeleteUserGrades"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.Int("grade_id", int(req.Id)),
	))
	defer span.End()

	// 删除用户等级
	err := s.deleteUserGrade(ctx, int(req.SiteId), int(req.Id))
	if err != nil {
		tracing.SetSpanError(span, err)
		return &v1.DeleteUserGradesRes{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	tracing.SetSpanAttributes(span, attribute.Bool("success", true))
	return &v1.DeleteUserGradesRes{
		Code:    200,
		Message: "删除成功",
	}, nil
}

// getUserGradeList 获取用户等级列表（内部方法）
func (s *sUser) getUserGradeList(ctx context.Context, siteId int) ([]*entity.UserGrade, error) {
	var grades []*entity.UserGrade

	// 使用默认数据库连接
	err := dao.UserGrade.Ctx(ctx).
		Where(dao.UserGrade.Columns().SiteId, siteId).
		Where(dao.UserGrade.Columns().Status, 1). // 只获取正常状态的等级
		OrderAsc(dao.UserGrade.Columns().Id).
		Scan(&grades)

	if err != nil {
		return nil, err
	}

	return grades, nil
}

// saveUserGrades 保存用户等级（内部方法）
func (s *sUser) saveUserGrades(ctx context.Context, siteId int, grades []*entity.UserGrade) error {
	// 使用默认数据库连接
	db := dao.UserGrade.Ctx(ctx)

	// 开启事务
	return db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, grade := range grades {
			grade.SiteId = siteId
			grade.Status = 1 // 默认状态为正常

			// 记录日志
			g.Log().Infof(ctx, "保存用户等级: ID=%d, Name=%s, BetLimit=%f, PaymentLimit=%f, MoneyLimit=%f",
				grade.Id, grade.Name, grade.BetLimit, grade.PaymentLimit, grade.MoneyLimit)

			if grade.Id > 0 {
				// 更新现有等级
				grade.UpdatedAt = gtime.Now()

				// 构建更新数据
				updateData := g.Map{
					"name":                   grade.Name,
					"rebate_percent_sports":  grade.RebatePercentSports,
					"rebate_percent_lottery": grade.RebatePercentLottery,
					"rebate_percent_live":    grade.RebatePercentLive,
					"rebate_percent_egame":   grade.RebatePercentEgame,
					"rebate_percent_poker":   grade.RebatePercentPoker,
					"payment_limit":          grade.PaymentLimit,
					"bet_limit":              grade.BetLimit,
					"money_limit":            grade.MoneyLimit,
					"status":                 grade.Status,
					"updated_at":             grade.UpdatedAt,
				}

				result, err := tx.Model(dao.UserGrade.Table()).
					Where(dao.UserGrade.Columns().Id, grade.Id).
					Where(dao.UserGrade.Columns().SiteId, siteId).
					Data(updateData).
					Update()
				if err != nil {
					g.Log().Errorf(ctx, "更新用户等级失败: %v", err)
					return err
				}

				rowsAffected, _ := result.RowsAffected()
				g.Log().Infof(ctx, "更新用户等级成功: ID=%d, 影响行数=%d", grade.Id, rowsAffected)
			} else {
				// 创建新等级
				grade.CreatedAt = gtime.Now()
				grade.UpdatedAt = gtime.Now()
				_, err := tx.Model(dao.UserGrade.Table()).Data(grade).Insert()
				if err != nil {
					g.Log().Errorf(ctx, "创建用户等级失败: %v", err)
					return err
				}
				g.Log().Infof(ctx, "创建用户等级成功: Name=%s", grade.Name)
			}
		}
		return nil
	})
}

// deleteUserGrade 删除用户等级（内部方法）
func (s *sUser) deleteUserGrade(ctx context.Context, siteId, gradeId int) error {
	// 使用默认数据库连接
	db := dao.UserGrade.DB()

	// 记录删除操作的日志
	g.Log().Infof(ctx, "开始删除用户等级: siteId=%d, gradeId=%d", siteId, gradeId)

	// 首先检查要删除的等级是否存在
	var existingGrade entity.UserGrade
	err := db.Model(dao.UserGrade.Table()).
		Where(dao.UserGrade.Columns().Id, gradeId).
		Where(dao.UserGrade.Columns().SiteId, siteId).
		Scan(&existingGrade)
	if err != nil {
		g.Log().Errorf(ctx, "查询等级失败: %v", err)
		if err.Error() == "sql: no rows in result set" {
			return fmt.Errorf("要删除的等级不存在")
		}
		return fmt.Errorf("查询等级信息失败: %v", err)
	}

	g.Log().Infof(ctx, "找到要删除的等级: %+v", existingGrade)

	// 简化版本：直接删除等级，暂时跳过业务逻辑检查
	result, err := db.Model(dao.UserGrade.Table()).
		Where(dao.UserGrade.Columns().Id, gradeId).
		Where(dao.UserGrade.Columns().SiteId, siteId).
		Delete()

	if err != nil {
		g.Log().Errorf(ctx, "删除等级失败: %v", err)
		return fmt.Errorf("删除等级失败: %v", err)
	}

	// 检查是否真的删除了记录
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		g.Log().Errorf(ctx, "获取删除结果失败: %v", err)
		return fmt.Errorf("获取删除结果失败: %v", err)
	}

	g.Log().Infof(ctx, "删除操作影响行数: %d", rowsAffected)

	if rowsAffected == 0 {
		return fmt.Errorf("没有删除任何记录，等级可能不存在")
	}

	g.Log().Infof(ctx, "成功删除用户等级: siteId=%d, gradeId=%d", siteId, gradeId)
	return nil
}

// getUserCountByGrade 获取指定等级的用户数量（内部方法）
func (s *sUser) getUserCountByGrade(ctx context.Context, siteId, gradeId int) (int, error) {
	// 使用默认数据库连接
	db := dao.UserGrade.DB()

	count, err := db.Model("user").
		Where("site_id", siteId).
		Where("grade_id", gradeId).
		Count()

	return count, err
}

// GetUserLoginLogs 获取用户登录日志
func (s *sUser) GetUserLoginLogs(ctx context.Context, req *v1.GetUserLoginLogsReq) (*v1.GetUserLoginLogsRes, error) {
	// 参数验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "user.GetUserLoginLogs", trace.WithAttributes(
		attribute.String("method", "GetUserLoginLogs"),
		attribute.String("username", req.Username),
		attribute.String("ip", req.Ip),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	// 获取站点ID

	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("点ID不存在")
	}
	siteCode, err := dao.Site.GetCodeById(ctx, siteId)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "未获取站点code")
		return nil, fmt.Errorf("获取站点CODE失败: %v", err)
	}

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("配置的站点代码: %s", siteCode))

	tracing.SetSpanAttributes(span, attribute.Int("site_id", int(siteId)))

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("开始查询用户登录日志，站点ID: %d, 页码: %d, 每页: %d", siteId, req.Page, req.Size))

	// 设置默认分页参数
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	// 构建查询条件
	query := dao.UserLoginLog.Ctx(ctx).Where("site_id = ?", siteId)

	// 添加筛选条件
	if req.Username != "" {
		query = query.Where("username = ?", req.Username)
	}
	if req.Ip != "" {
		query = query.Where("login_ip = ?", req.Ip)
	}
	if req.StartTime != "" {
		query = query.Where("login_time >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("login_time <= ?", req.EndTime)
	}

	// 获取总数
	ctx, countSpan := tracing.StartSpan(ctx, "db.query.user_login_log_count", trace.WithAttributes(
		attribute.String("db.operation", "count"),
		attribute.String("db.table", "user_login_log"),
	))
	count, err := query.Count()
	countSpan.End()
	if err != nil {
		middleware.LogWithTrace(ctx, "error", fmt.Sprintf("查询用户登录日志总数失败: %v", err))
		return nil, fmt.Errorf("查询用户登录日志总数失败: %v", err)
	}

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("用户登录日志总数: %d", count))
	tracing.SetSpanAttributes(span, attribute.Int("total_count", count))

	// 分页查询
	offset := (page - 1) * size
	ctx, listSpan := tracing.StartSpan(ctx, "db.query.user_login_log_list", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "user_login_log"),
		attribute.Int("limit", int(size)),
		attribute.Int("offset", int(offset)),
	))

	var logs []*entity.UserLoginLog
	err = query.Order("login_time DESC, created_at DESC, id DESC").
		Limit(int(size)).
		Offset(int(offset)).
		Scan(&logs)
	listSpan.End()

	if err != nil {
		middleware.LogWithTrace(ctx, "error", fmt.Sprintf("查询用户登录日志列表失败: %v", err))
		return nil, fmt.Errorf("查询用户登录日志列表失败: %v", err)
	}

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("查询到 %d 条用户登录日志", len(logs)))

	// 转换数据格式
	var list []*v1.UserLoginLogInfo
	for _, log := range logs {
		deviceName := s.getDeviceName(log.Device)

		logInfo := &v1.UserLoginLogInfo{
			Id:           int32(log.Id),
			UserId:       int32(log.UserId),
			Username:     log.Username,
			RefererUrl:   log.RefererUrl,
			LoginUrl:     log.LoginUrl,
			LoginTime:    log.LoginTime.String(),
			LoginIp:      log.LoginIp,
			LoginAddress: log.LoginAddress,
			Os:           log.Os,
			Network:      log.Network,
			Screen:       log.Screen,
			Browser:      log.Browser,
			Device:       deviceName,
			IsRobot:      int32(log.IsRobot),
			CreatedAt:    log.CreatedAt.String(),
		}
		list = append(list, logInfo)
	}

	return &v1.GetUserLoginLogsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// IncreasePayTimes 增加用户充值次数
func (s *sUser) IncreasePayTimes(ctx context.Context, req *v1.IncreasePayTimesReq) (*v1.IncreasePayTimesRes, error) {
	if req == nil || req.UserId <= 0 {
		return &v1.IncreasePayTimesRes{Success: false, Message: "用户ID不能为空"}, nil
	}

	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists || siteId == 0 {
		if req.SiteId <= 0 {
			return &v1.IncreasePayTimesRes{Success: false, Message: "站点ID不存在"}, nil
		}
		siteId = uint(req.SiteId)
	}

	var (
		oldPayTimes int
		newPayTimes int
		inviterId   int
		siteCode    string
	)

	err := dao.User.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user entity.User
		if err := dao.User.Ctx(ctx).TX(tx).
			Where("site_id", siteId).
			Where("id", req.UserId).
			LockUpdate().
			Scan(&user); err != nil {
			return err
		}
		if user.Id == 0 {
			return fmt.Errorf("用户不存在")
		}

		oldPayTimes = user.PayTimes
		newPayTimes = oldPayTimes + 1
		inviterId = user.InviterId

		_, err := dao.User.Ctx(ctx).TX(tx).
			Where("id", user.Id).
			Update(g.Map{
				"pay_times":  newPayTimes,
				"updated_at": gtime.Now(),
			})
		if err != nil {
			return err
		}

		// 解析站点代码（优先从context取）
		if ctxSiteCode, ok := middleware.GetSiteCodeFromContext(ctx); ok {
			siteCode = ctxSiteCode
		}
		if siteCode == "" {
			siteCode = registry.GetDBConfigManagerV2().GetSiteCodeBySiteID(int64(siteId))
		}

		// 同步用户到 ClickHouse 的 Outbox
		user.PayTimes = newPayTimes
		user.UpdatedAt = gtime.Now()
		if err := logic_mq_outbox.MqOutbox().CreateSyncUserToClickhouseOutbox(ctx, tx, int64(siteId), &user, siteCode); err != nil {
			return fmt.Errorf("创建同步用户到ClickHouse Outbox记录失败: %w", err)
		}

		return nil
	})
	if err != nil {
		middleware.LogWithTrace(ctx, "error", fmt.Sprintf("更新用户充值次数失败: user_id=%d, err=%v", req.UserId, err))
		return &v1.IncreasePayTimesRes{Success: false, Message: "更新充值次数失败"}, nil
	}

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("更新用户充值次数成功: user_id=%d, old=%d, new=%d", req.UserId, oldPayTimes, newPayTimes))
	return &v1.IncreasePayTimesRes{
		Success:     true,
		Message:     "更新充值次数成功",
		OldPayTimes: int32(oldPayTimes),
		NewPayTimes: int32(newPayTimes),
		InviterId:   int32(inviterId),
	}, nil
}

// getDeviceName 获取设备名称
func (s *sUser) getDeviceName(device int) string {
	switch device {
	case 1:
		return "电脑"
	case 2:
		return "手机"
	case 3:
		return "平板"
	default:
		return "未知"
	}
}
