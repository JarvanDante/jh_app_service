package message

import (
	"context"
	"fmt"
	"jh_app_service/api/backend/message/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/service/backend"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
)

type (
	sMessage struct{}
)

var (
	placeholderPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)
)

func init() {
	backend.RegisterMessage(&sMessage{})
}

// GetMessageList 获取消息列表
func (s *sMessage) GetMessageList(ctx context.Context, req *v1.GetMessageListReq) (*v1.GetMessageListRes, error) {
	middleware.LogWithTrace(ctx, "info", "获取消息列表请求 - Page: %d, Size: %d", req.Page, req.Size)

	siteId := int32(1)
	if sid, ok := middleware.GetSiteIdFromContext(ctx); ok {
		siteId = int32(sid)
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 50
	}

	// 获取消息列表
	type messageRow struct {
		Id        uint        `json:"id"`
		EventCode string      `json:"event_code"`
		Title     string      `json:"title"`
		Content   string      `json:"content"`
		Status    int         `json:"status"`
		CreatedAt *gtime.Time `json:"created_at"`
		UpdatedAt *gtime.Time `json:"updated_at"`
	}
	var messages []*messageRow
	total, err := dao.Message.Ctx(ctx).Where("site_id", siteId).Count()
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "获取消息总数失败: %v", err)
		return nil, err
	}

	// 分页查询
	offset := int((page - 1) * size)
	err = dao.Message.Ctx(ctx).
		Where("site_id", siteId).
		Fields("id, event_code, title, content, status, created_at, updated_at").
		Order("created_at DESC").
		Limit(offset, int(size)).
		Scan(&messages)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "获取消息列表失败: %v", err)
		return nil, err
	}

	// 转换为响应格式
	messageList := make([]*v1.MessageItem, 0, len(messages))
	for _, msg := range messages {
		createdAt := ""
		if msg.CreatedAt != nil {
			createdAt = msg.CreatedAt.Format("2006-01-02 15:04:05")
		}
		updatedAt := ""
		if msg.UpdatedAt != nil {
			updatedAt = msg.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		messageList = append(messageList, &v1.MessageItem{
			Id:        int32(msg.Id),
			EventCode: msg.EventCode,
			Title:     msg.Title,
			Content:   msg.Content,
			Status:    int32(msg.Status),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	// 获取状态列表
	statusList := []*v1.StatusItem{
		{Value: 1, Name: "启用"},
		{Value: 0, Name: "关闭"},
	}

	middleware.LogWithTrace(ctx, "info", "获取消息列表成功 - 总数: %d, 当前页: %d", total, page)

	return &v1.GetMessageListRes{
		MessageList: messageList,
		GradeList:   []*v1.GradeItem{},
		LevelList:   []*v1.LevelItem{},
		StatusList:  statusList,
		Total:       int32(total),
		CurrentPage: page,
		PerPage:     size,
	}, nil
}

// CreateMessage 创建消息
func (s *sMessage) CreateMessage(ctx context.Context, req *v1.CreateMessageReq) (*v1.CreateMessageRes, error) {
	middleware.LogWithTrace(ctx, "info", "创建消息请求 - Title: %s", req.Title)

	siteId := int32(1)
	if sid, ok := middleware.GetSiteIdFromContext(ctx); ok {
		siteId = int32(sid)
	}
	adminId := int32(1)
	if aid, ok := middleware.GetAdminIdFromContext(ctx); ok {
		adminId = int32(aid)
	}

	// 验证参数
	if req.EventCode == "" {
		middleware.LogWithTrace(ctx, "error", "业务触发编码不能为空")
		return nil, fmt.Errorf("业务触发编码不能为空")
	}
	if req.Title == "" || len(req.Title) < 2 || len(req.Title) > 255 {
		middleware.LogWithTrace(ctx, "error", "标题长度必须在2-255字符之间")
		return nil, fmt.Errorf("标题长度必须在2-255字符之间")
	}
	if req.Content == "" || len(req.Content) < 2 || len(req.Content) > 5000 {
		middleware.LogWithTrace(ctx, "error", "内容长度必须在2-5000字符之间")
		return nil, fmt.Errorf("内容长度必须在2-5000字符之间")
	}
	if req.Status != 0 && req.Status != 1 {
		middleware.LogWithTrace(ctx, "error", "状态只能为0或1")
		return nil, fmt.Errorf("状态只能为0或1")
	}

	// 创建消息记录
	message := g.Map{
		"site_id":    siteId,
		"admin_id":   adminId,
		"event_code": req.EventCode,
		"title":      req.Title,
		"content":    req.Content,
		"status":     req.Status,
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	}

	// 保存消息
	result, err := dao.Message.Ctx(ctx).Data(message).Insert()
	if err != nil {
		if gstr.Contains(err.Error(), "Duplicate entry") {
			middleware.LogWithTrace(ctx, "error", "业务触发编码已存在")
			return nil, fmt.Errorf("业务触发编码已存在")
		}
		middleware.LogWithTrace(ctx, "error", "创建消息失败: %v", err)
		return nil, err
	}

	messageId, err := result.LastInsertId()
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "获取消息ID失败: %v", err)
		return nil, err
	}

	// 根据类型发送消息给用户（兼容旧逻辑，未设置Type则不触发）
	if req.Type > 0 {
		err = s.sendMessageToUsers(ctx, messageId, req)
		if err != nil {
			g.Log().Errorf(ctx, "发送消息给用户失败: %v", err)
			// 不返回错误，消息已创建成功
		}
	}

	middleware.LogWithTrace(ctx, "info", "创建消息成功 - Title: %s, ID: %d", req.Title, messageId)

	return &v1.CreateMessageRes{}, nil
}

// UpdateMessage 更新消息
func (s *sMessage) UpdateMessage(ctx context.Context, req *v1.UpdateMessageReq) (*v1.UpdateMessageRes, error) {
	middleware.LogWithTrace(ctx, "info", "更新消息请求 - ID: %d", req.Id)

	siteId := int32(1)
	if sid, ok := middleware.GetSiteIdFromContext(ctx); ok {
		siteId = int32(sid)
	}

	if req.Id <= 0 {
		return nil, fmt.Errorf("消息ID不能为空")
	}
	if req.Status != 0 && req.Status != 1 {
		return nil, fmt.Errorf("状态只能为0或1")
	}

	updateData := g.Map{}
	if req.EventCode != "" {
		updateData["event_code"] = req.EventCode
	}
	if req.Title != "" {
		if len(req.Title) < 2 || len(req.Title) > 255 {
			return nil, fmt.Errorf("标题长度必须在2-255字符之间")
		}
		updateData["title"] = req.Title
	}
	if req.Content != "" {
		if len(req.Content) < 2 || len(req.Content) > 5000 {
			return nil, fmt.Errorf("内容长度必须在2-5000字符之间")
		}
		updateData["content"] = req.Content
	}
	updateData["status"] = req.Status
	updateData["updated_at"] = gtime.Now()

	if len(updateData) == 0 {
		return nil, fmt.Errorf("没有需要更新的字段")
	}

	_, err := dao.Message.Ctx(ctx).
		Where("site_id", siteId).
		Where("id", req.Id).
		Data(updateData).
		Update()
	if err != nil {
		if gstr.Contains(err.Error(), "Duplicate entry") {
			return nil, fmt.Errorf("业务触发编码已存在")
		}
		return nil, err
	}

	return &v1.UpdateMessageRes{}, nil
}

// DeleteMessage 删除消息
func (s *sMessage) DeleteMessage(ctx context.Context, req *v1.DeleteMessageReq) (*v1.DeleteMessageRes, error) {
	middleware.LogWithTrace(ctx, "info", "删除消息请求 - ID: %d", req.Id)

	siteId := int32(1)
	if sid, ok := middleware.GetSiteIdFromContext(ctx); ok {
		siteId = int32(sid)
	}

	if req.Id <= 0 {
		return nil, fmt.Errorf("消息ID不能为空")
	}

	_, err := dao.Message.Ctx(ctx).
		Where("site_id", siteId).
		Where("id", req.Id).
		Delete()
	if err != nil {
		return nil, err
	}

	return &v1.DeleteMessageRes{}, nil
}

// GetUserMessages 获取用户消息
func (s *sMessage) GetUserMessages(ctx context.Context, req *v1.GetUserMessagesReq) (*v1.GetUserMessagesRes, error) {
	middleware.LogWithTrace(ctx, "info", "获取用户消息请求 - UserId: %d", req.UserId)

	// 默认站点ID为1
	siteId := int32(1)
	if sid, ok := middleware.GetSiteIdFromContext(ctx); ok && sid > 0 {
		siteId = int32(sid)
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 20
	}

	// 定义查询结果结构体
	type UserMessageResult struct {
		Id        uint        `json:"id"`
		Title     string      `json:"title"`
		Content   string      `json:"content"`
		IsRead    int         `json:"is_read"`
		CreatedAt *gtime.Time `json:"created_at"`
	}

	// 获取用户消息列表
	var userMessages []*UserMessageResult
	total, err := dao.UserMessage.Ctx(ctx).
		Where("user_message.site_id", siteId).
		Where("user_message.user_id", req.UserId).
		Count()
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "获取用户消息总数失败: %v", err)
		return nil, err
	}

	// 分页查询
	offset := int((page - 1) * size)
	err = dao.UserMessage.Ctx(ctx).
		Where("user_message.site_id", siteId).
		Where("user_message.user_id", req.UserId).
		Fields("user_message.id, user_message.title, user_message.content, user_message.status as is_read, user_message.created_at").
		Order("user_message.created_at DESC").
		Limit(offset, int(size)).
		Scan(&userMessages)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "获取用户消息失败: %v", err)
		return nil, err
	}

	// 转换为响应格式
	messageList := make([]*v1.UserMessageItem, 0, len(userMessages))
	for _, msg := range userMessages {
		messageList = append(messageList, &v1.UserMessageItem{
			Id:        int32(msg.Id),
			Title:     msg.Title,
			Content:   msg.Content,
			IsRead:    int32(msg.IsRead),
			CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// 获取未读消息数量
	unreadCount, err := dao.UserMessage.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", req.UserId).
		Where("status", 0). // 0表示未读
		Count()
	if err != nil {
		g.Log().Errorf(ctx, "获取未读消息数量失败: %v", err)
		unreadCount = 0
	}

	middleware.LogWithTrace(ctx, "info", "获取用户消息成功 - UserId: %d, 总数: %d", req.UserId, total)

	return &v1.GetUserMessagesRes{
		List:               messageList,
		Total:              int32(total),
		CurrentPage:        page,
		PerPage:            size,
		UnreadMessageCount: int32(unreadCount),
	}, nil
}

// GetUnreadCount 获取未读数量
func (s *sMessage) GetUnreadCount(ctx context.Context, req *v1.GetUnreadCountReq) (*v1.GetUnreadCountRes, error) {
	middleware.LogWithTrace(ctx, "info", "获取未读数量请求 - UserId: %d", req.UserId)

	// 默认站点ID为1
	siteId := int32(1)
	if sid, ok := middleware.GetSiteIdFromContext(ctx); ok && sid > 0 {
		siteId = int32(sid)
	}

	count, err := dao.UserMessage.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", req.UserId).
		Where("status", 0). // 0表示未读
		Count()
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "获取未读消息数量失败: %v", err)
		return nil, err
	}

	middleware.LogWithTrace(ctx, "info", "获取未读数量成功 - UserId: %d, Count: %d", req.UserId, count)

	return &v1.GetUnreadCountRes{
		Count: int32(count),
	}, nil
}

// ReadMessage 读取消息
func (s *sMessage) ReadMessage(ctx context.Context, req *v1.ReadMessageReq) (*v1.ReadMessageRes, error) {
	middleware.LogWithTrace(ctx, "info", "读取消息请求 - UserId: %d, MessageId: %d", req.UserId, req.MessageId)

	// 默认站点ID为1
	siteId := int32(1)
	if sid, ok := middleware.GetSiteIdFromContext(ctx); ok && sid > 0 {
		siteId = int32(sid)
	}

	_, err := dao.UserMessage.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", req.UserId).
		Where("id", req.MessageId).
		Data(g.Map{"status": 1}). // 1表示已读
		Update()
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "标记消息已读失败: %v", err)
		return nil, err
	}

	middleware.LogWithTrace(ctx, "info", "读取消息成功 - UserId: %d, MessageId: %d", req.UserId, req.MessageId)

	return &v1.ReadMessageRes{}, nil
}

// sendMessageToUsers 根据类型发送消息给用户
func (s *sMessage) sendMessageToUsers(ctx context.Context, messageId int64, req *v1.CreateMessageReq) error {
	// 默认站点ID为1
	siteId := int32(1)

	baseData := g.Map{
		"site_id":    siteId,
		"message_id": int(messageId),
		"event_code": req.EventCode,
		"biz_id":     "",
		"title":      req.Title,
		"content":    req.Content,
		"status":     0,
		"created_at": gtime.Now(),
	}
	if req.Amount > 0 {
		baseData["amount"] = req.Amount
	}

	switch req.Type {
	case 1: // 指定用户
		// 解析用户名列表，创建用户消息记录
		// 这里简化处理，实际需要根据用户名查询用户ID
		userIds := []int{1, 2, 3} // 写死数据，后续快照用
		for _, userId := range userIds {
			data := g.Map{
				"site_id":    baseData["site_id"],
				"message_id": baseData["message_id"],
				"event_code": baseData["event_code"],
				"biz_id":     baseData["biz_id"],
				"title":      baseData["title"],
				"content":    baseData["content"],
				"status":     baseData["status"],
				"created_at": baseData["created_at"],
				"user_id":    userId,
			}
			if amount, ok := baseData["amount"]; ok {
				data["amount"] = amount
			}
			_, err := dao.UserMessage.Ctx(ctx).Data(data).Insert()
			if err != nil {
				g.Log().Errorf(ctx, "创建用户消息记录失败: %v", err)
			}
		}
	case 2: // 等级
		// 根据等级查询用户，创建消息记录
		// 写死数据，后续快照用
		userIds := []int{1, 2} // 假设等级1的用户
		for _, userId := range userIds {
			data := g.Map{
				"site_id":    baseData["site_id"],
				"message_id": baseData["message_id"],
				"event_code": baseData["event_code"],
				"biz_id":     baseData["biz_id"],
				"title":      baseData["title"],
				"content":    baseData["content"],
				"status":     baseData["status"],
				"created_at": baseData["created_at"],
				"user_id":    userId,
			}
			if amount, ok := baseData["amount"]; ok {
				data["amount"] = amount
			}
			_, err := dao.UserMessage.Ctx(ctx).Data(data).Insert()
			if err != nil {
				g.Log().Errorf(ctx, "创建用户消息记录失败: %v", err)
			}
		}
	case 3: // 层级
		// 根据层级查询用户，创建消息记录
		// 写死数据，后续快照用
		userIds := []int{3, 4} // 假设层级1的用户
		for _, userId := range userIds {
			data := g.Map{
				"site_id":    baseData["site_id"],
				"message_id": baseData["message_id"],
				"event_code": baseData["event_code"],
				"biz_id":     baseData["biz_id"],
				"title":      baseData["title"],
				"content":    baseData["content"],
				"status":     baseData["status"],
				"created_at": baseData["created_at"],
				"user_id":    userId,
			}
			if amount, ok := baseData["amount"]; ok {
				data["amount"] = amount
			}
			_, err := dao.UserMessage.Ctx(ctx).Data(data).Insert()
			if err != nil {
				g.Log().Errorf(ctx, "创建用户消息记录失败: %v", err)
			}
		}
	case 4: // 状态
		// 根据状态查询用户，创建消息记录
		// 写死数据，后续快照用
		userIds := []int{1, 2, 3, 4} // 假设在线用户
		for _, userId := range userIds {
			data := g.Map{
				"site_id":    baseData["site_id"],
				"message_id": baseData["message_id"],
				"event_code": baseData["event_code"],
				"biz_id":     baseData["biz_id"],
				"title":      baseData["title"],
				"content":    baseData["content"],
				"status":     baseData["status"],
				"created_at": baseData["created_at"],
				"user_id":    userId,
			}
			if amount, ok := baseData["amount"]; ok {
				data["amount"] = amount
			}
			_, err := dao.UserMessage.Ctx(ctx).Data(data).Insert()
			if err != nil {
				g.Log().Errorf(ctx, "创建用户消息记录失败: %v", err)
			}
		}
	}

	return nil
}

// SendUserMessage 根据模板发送用户消息
func (s *sMessage) SendUserMessage(ctx context.Context, req *v1.SendUserMessageReq) (*v1.SendUserMessageRes, error) {
	if req == nil {
		return &v1.SendUserMessageRes{Success: false, Message: "请求参数不能为空"}, nil
	}

	siteId := int32(0)
	if sid, ok := middleware.GetSiteIdFromContext(ctx); ok && sid > 0 {
		siteId = int32(sid)
	} else if req.SiteId > 0 {
		siteId = req.SiteId
	}
	if siteId <= 0 {
		return &v1.SendUserMessageRes{Success: false, Message: "未获取站点ID"}, nil
	}
	if req.UserId <= 0 {
		return &v1.SendUserMessageRes{Success: false, Message: "用户ID不能为空"}, nil
	}
	if strings.TrimSpace(req.EventCode) == "" {
		return &v1.SendUserMessageRes{Success: false, Message: "业务触发编码不能为空"}, nil
	}
	if strings.TrimSpace(req.BizId) == "" {
		return &v1.SendUserMessageRes{Success: false, Message: "业务流水号不能为空"}, nil
	}

	type messageTpl struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  int    `json:"status"`
	}
	var tpl messageTpl
	if err := dao.Message.Ctx(ctx).
		Where("site_id", siteId).
		Where("event_code", strings.TrimSpace(req.EventCode)).
		Where("status", 1).
		Fields("id, title, content, status").
		Scan(&tpl); err != nil {
		return &v1.SendUserMessageRes{Success: false, Message: "查询消息模板失败"}, nil
	}
	if tpl.Id <= 0 {
		return &v1.SendUserMessageRes{Success: false, Message: "消息模板不存在或未启用"}, nil
	}

	vars := buildTemplateVars(req)
	title := renderTemplate(tpl.Title, vars)
	content := renderTemplate(tpl.Content, vars)

	data := g.Map{
		"site_id":    siteId,
		"user_id":    req.UserId,
		"message_id": tpl.Id,
		"event_code": strings.TrimSpace(req.EventCode),
		"biz_id":     strings.TrimSpace(req.BizId),
		"title":      title,
		"content":    content,
		"status":     0,
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	}
	if req.Amount > 0 {
		data["amount"] = req.Amount
	}
	if strings.TrimSpace(req.GradeName) != "" {
		data["grade_name"] = strings.TrimSpace(req.GradeName)
	}

	if _, err := dao.UserMessage.Ctx(ctx).Data(data).Insert(); err != nil {
		if gstr.Contains(err.Error(), "Duplicate entry") || gstr.Contains(strings.ToLower(err.Error()), "duplicate") {
			return &v1.SendUserMessageRes{Success: true, Message: "消息已存在"}, nil
		}
		return &v1.SendUserMessageRes{Success: false, Message: "发送消息失败"}, nil
	}

	return &v1.SendUserMessageRes{Success: true, Message: "发送成功"}, nil
}

func buildTemplateVars(req *v1.SendUserMessageReq) map[string]string {
	vars := make(map[string]string)
	for k, v := range req.GetVars() {
		key := strings.ToLower(strings.TrimSpace(k))
		if key == "" {
			continue
		}
		vars[key] = v
	}
	if req.Amount > 0 {
		vars["amount"] = fmt.Sprintf("%.2f", req.Amount)
	}
	if strings.TrimSpace(req.GradeName) != "" {
		grade := strings.TrimSpace(req.GradeName)
		vars["grade_name"] = grade
		vars["gradename"] = grade
	}
	return vars
}

func renderTemplate(text string, vars map[string]string) string {
	if text == "" || len(vars) == 0 {
		return text
	}
	return placeholderPattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := placeholderPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		key := strings.ToLower(strings.TrimSpace(parts[1]))
		if val, ok := vars[key]; ok {
			return val
		}
		return ""
	})
}
