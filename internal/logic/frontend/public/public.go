package public

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"

	v1 "jh_app_service/api/frontend/public/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/mq"
	"jh_app_service/internal/registry"
	"jh_app_service/internal/service/frontend"
	"jh_app_service/internal/tracing"
	"jh_app_service/internal/util"
)

type sPublic struct{}

func New() *sPublic {
	return &sPublic{}
}

func init() {
	// 注册服务
	frontend.RegisterPublic(New())
}

// QuickRegister 快速注册
func (s *sPublic) QuickRegister(ctx context.Context, req *v1.QuickRegisterReq) (res *v1.QuickRegisterRes, err error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "user.QuickRegister", trace.WithAttributes(
		attribute.String("method", "QuickRegister"),
	))
	defer span.End()

	// 获取站点ID (这里需要根据实际情况获取，可能从上下文或配置中获取)
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "site_id_not_found")
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	// 生成随机用户名（12位M）
	username := grand.S(12)

	password := username

	// 使用bcrypt加密密码
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		g.Log().Error(ctx, "密码加密失败:", err)
		return nil, fmt.Errorf("注册失败")
	}

	// 获取客户端IP
	// clientIp := g.RequestFromCtx(ctx).GetClientIp()
	clientIp := "127.0.0.1" // TODO: 从context获取真实IP

	// 获取站点配置
	var siteConfig *entity.SiteConfig
	err = dao.SiteConfig.Ctx(ctx).Where("site_id", siteId).Scan(&siteConfig)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &entity.User{
		SiteId:       int(siteId),
		GradeId:      siteConfig.DefaultGradeId, // TODO: 获取默认等级
		LevelId:      siteConfig.DefaultLevelId, // TODO: 获取默认层级
		AgentId:      siteConfig.DefaultAgentId, // TODO: 根据代理码获取代理ID
		Username:     username,
		Password:     hashedPassword,
		PayPassword:  "",
		Realname:     "",
		Mobile:       "",
		Email:        "",
		SafeQuestion: "",
		SafeAnswer:   "",
		RegisterIp:   clientIp,
		RegisterTime: gtime.Now(),
		RegisterUrl:  req.Domain,
		Status:       1,
	}

	// 保存用户并获取ID
	result, err := dao.User.Ctx(ctx).Insert(user)
	if err != nil {
		g.Log().Error(ctx, "快速注册失败:", err)
		return nil, fmt.Errorf("注册失败")
	}

	// 获取插入的用户ID
	userId, err := result.LastInsertId()
	if err != nil {
		g.Log().Error(ctx, "获取用户ID失败:", err)
		return nil, fmt.Errorf("注册失败")
	}
	user.Id = uint(userId)

	// 异步调用balance服务初始化余额
	go func() {
		// 获取站点代码
		siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
		if siteCode == "" {
			g.Log().Error(ctx, "获取站点代码失败，无法初始化余额")
			return
		}

		// 从Consul KV获取站点配置的currency
		dbManager := registry.GetDBConfigManagerV2()
		siteConfig, err := dbManager.GetSiteConfig(siteCode)
		if err != nil {
			g.Log().Errorf(ctx, "获取站点配置失败: %v", err)
			return
		}

		currency := siteConfig.Base.Currency
		if currency == "" {
			currency = "CNY" // 默认人民币
		}

		// 发送RabbitMQ消息
		msg := &mq.InitBalanceUserMessage{
			SiteID:   int32(siteId),
			UserID:   int32(userId),
			Username: username,
			Currency: currency,
		}

		if err := mq.PublishInitBalanceUser(context.Background(), msg); err != nil {
			g.Log().Errorf(ctx, "发送初始化余额消息失败: %v", err)
		}
	}()

	// TODO: 以下操作需要调用其他微服务或异步处理
	// 2. 代理注册奖金 - 调用 balance_service
	// 3. 风险预警 - 异步处理
	// 4. 自定义资源绑定 - 异步处理

	// 生成JWT Token和Refresh Token
	token, refreshToken, expiresIn, refreshExpiresIn, err := s.generateJWTTokenUser(user)
	if err != nil {
		g.Log().Error(ctx, "生成token失败:", err)
		return nil, fmt.Errorf("注册失败")
	}

	res = &v1.QuickRegisterRes{
		Token:            token,
		Username:         username,
		RefreshToken:     refreshToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}

	return res, nil
}

// generateJWTTokenUser 生成JWT token和refresh token
func (s *sPublic) generateJWTTokenUser(user *entity.User) (token string, refreshToken string, expiresIn int64, refreshExpiresIn int64, err error) {
	// 从配置文件获取JWT密钥
	jwtSecret := g.Cfg().MustGet(context.Background(), "jwt.secret").String()
	if jwtSecret == "" {
		return "", "", 0, 0, fmt.Errorf("JWT secret not configured")
	}

	now := time.Now()

	// Token过期时间：1小时
	expiresIn = 3600
	tokenExpireTime := now.Add(time.Duration(expiresIn) * time.Second)

	// Refresh token过期时间：30天
	refreshExpiresIn = 2592000 // 30 * 24 * 3600
	refreshTokenExpireTime := now.Add(time.Duration(refreshExpiresIn) * time.Second)

	// 生成唯一的token ID（jti）
	tokenJti := fmt.Sprintf("%d_%d_%d", user.Id, user.SiteId, now.UnixNano())
	refreshJti := fmt.Sprintf("refresh_%d_%d_%d", user.Id, user.SiteId, now.UnixNano())

	// 创建access token
	token_claims := jwt.MapClaims{
		"user_id":  user.Id,
		"admin_id": 0,
		"username": user.Username,
		"site_id":  user.SiteId,
		"exp":      tokenExpireTime.Unix(),
		"iat":      now.Unix(),
		"type":     "access", // 标记为access token
		"jti":      tokenJti, // JWT ID，用于黑名单
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token_claims)
	token, err = accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", 0, 0, err
	}

	// 创建refresh token（包含关联的access token jti）
	refresh_claims := jwt.MapClaims{
		"user_id":          user.Id,
		"username":         user.Username,
		"site_id":          user.SiteId,
		"exp":              refreshTokenExpireTime.Unix(),
		"iat":              now.Unix(),
		"type":             "refresh",  // 标记为refresh token
		"jti":              refreshJti, // JWT ID，用于黑名单
		"access_token_jti": tokenJti,   // 关联的access token jti
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refresh_claims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", 0, 0, err
	}

	return token, refreshToken, expiresIn, refreshExpiresIn, nil
}

// generateGamePassword 生成游戏密码
func (s *sPublic) generateGamePassword() string {
	// 生成8位随机密码
	return grand.S(8)
}

// getRegisterDevice 获取注册设备
func (s *sPublic) getRegisterDevice(ctx context.Context) string {
	// TODO: 从User-Agent解析设备类型
	return "web"
}

// hashPassword 使用bcrypt加密密码
func (s *sPublic) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// Login 用户登录
func (s *sPublic) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.Login", trace.WithAttributes(
		attribute.String("method", "Login"),
		attribute.String("username", req.Username),
	))
	defer span.End()

	middleware.LogWithTrace(ctx, "info", "开始登录: username=%s, code=%s, time=%s", req.Username, req.Code, req.Time)

	// 获取站点ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "site_id_not_found")
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	middleware.LogWithTrace(ctx, "info", "站点ID: %d", siteId)

	// 验证验证码
	if req.Code != "" && req.Time != "" {
		clientIp := util.GetClientIP(ctx)

		// 从context获取站点代码
		siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
		if siteCode == "" {
			siteCode = "default"
		}

		cacheKey := fmt.Sprintf("%s:captcha:%s:%s", siteCode, clientIp, req.Time)
		middleware.LogWithTrace(ctx, "info", "验证验证码: key=%s, code=%s, ip=%s, siteCode=%s", cacheKey, req.Code, clientIp, siteCode)

		// 从Redis获取验证码
		cachedCode, err := g.Redis().Get(ctx, cacheKey)
		if err != nil {
			middleware.LogWithTrace(ctx, "error", "Redis获取验证码失败: %v", err)
			return nil, fmt.Errorf("验证码验证失败")
		}

		if cachedCode.IsNil() {
			middleware.LogWithTrace(ctx, "error", "验证码不存在或已过期: %s", cacheKey)
			return nil, fmt.Errorf("验证码已过期，请重新获取")
		}

		// 验证码不区分大小写
		if !strings.EqualFold(cachedCode.String(), req.Code) {
			middleware.LogWithTrace(ctx, "error", "验证码错误: 输入=%s, 期望=%s", req.Code, cachedCode.String())
			return nil, fmt.Errorf("验证码错误")
		}

		// 验证成功后删除验证码（一次性使用）
		g.Redis().Del(ctx, cacheKey)
		middleware.LogWithTrace(ctx, "info", "验证码验证成功并已删除: %s", cacheKey)
	}

	// 查询用户
	var user *entity.User
	err = dao.User.Ctx(ctx).Where("username", req.Username).
		Where("site_id", siteId).
		Where("status", 1).
		Scan(&user)

	if err != nil || user == nil {
		middleware.LogWithTrace(ctx, "error", "用户不存在: %s", req.Username)
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "密码错误: %s", req.Username)
		// TODO: 记录登录失败次数
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 更新最后登录信息
	clientIp := util.GetClientIP(ctx)
	_, err = dao.User.Ctx(ctx).Where("id", user.Id).Update(map[string]interface{}{
		"last_login_ip":   clientIp,
		"last_login_time": gtime.Now(),
	})

	// TODO: 记录登录日志
	// TODO: 清空登录失败次数

	// 生成JWT Token和Refresh Token
	token, refreshToken, expiresIn, refreshExpiresIn, err := s.generateJWTTokenUser(user)
	if err != nil {
		g.Log().Error(ctx, "生成token失败:", err)
		return nil, fmt.Errorf("登录失败")
	}

	res = &v1.LoginRes{
		Token:            token,
		RefreshToken:     refreshToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}

	return res, nil
}

// GenerateCaptcha 生成验证码
func (s *sPublic) GenerateCaptcha(ctx context.Context, req *v1.GenerateCaptchaReq) (res *v1.GenerateCaptchaRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "captcha.Generate", trace.WithAttributes(
		attribute.String("method", "GenerateCaptcha"),
	))
	defer span.End()

	// 生成4位验证码（仅数字）
	code := fmt.Sprintf("%04d", grand.N(0, 9999))
	middleware.LogWithTrace(ctx, "info", "生成验证码: %s", code)

	// 生成验证码图片（先生成图片，确保这部分没问题）
	middleware.LogWithTrace(ctx, "info", "开始生成验证码图片...")
	imageData, err := s.generateCaptchaImage(code)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "生成验证码图片失败: %v", err)
		return nil, fmt.Errorf("生成验证码失败: %v", err)
	}

	if len(imageData) == 0 {
		middleware.LogWithTrace(ctx, "error", "生成的验证码图片数据为空")
		return nil, fmt.Errorf("生成验证码失败: 图片数据为空")
	}

	middleware.LogWithTrace(ctx, "info", "验证码图片生成成功，大小: %d bytes", len(imageData))

	// 将验证码存储到缓存（5分钟有效）
	// 使用defer recover防止Redis panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				middleware.LogWithTrace(ctx, "error", "Redis操作panic: %v", r)
			}
		}()

		clientIp := util.GetClientIP(ctx)

		// 从context获取站点代码
		siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
		if siteCode == "" {
			siteCode = "default"
		}

		cacheKey := fmt.Sprintf("%s:captcha:%s:%s", siteCode, clientIp, req.Time)

		if err := g.Redis().SetEX(ctx, cacheKey, code, int64(300)); err != nil {
			middleware.LogWithTrace(ctx, "warn", "验证码缓存失败: %v", err)
		} else {
			middleware.LogWithTrace(ctx, "info", "验证码已缓存: %s = %s (IP: %s)", cacheKey, code, clientIp)
		}
	}()

	middleware.LogWithTrace(ctx, "info", "开始构建响应对象...")

	res = &v1.GenerateCaptchaRes{}
	res.Image = imageData
	res.ContentType = "image/png"

	middleware.LogWithTrace(ctx, "info", "准备返回验证码响应，图片大小: %d bytes, content_type: %s", len(res.Image), res.ContentType)

	return res, nil
}

// generateCaptchaImage 生成验证码图片
func (s *sPublic) generateCaptchaImage(code string) ([]byte, error) {
	// 创建图片 (宽150, 高50) - 增大尺寸
	width := 150
	height := 50

	// 创建一个新的RGBA图像
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景色（浅灰色）
	bgColor := color.RGBA{240, 240, 240, 255}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, bgColor)
		}
	}

	// 添加干扰线（减少到10条，避免太乱）
	// 添加干扰线（5条，更细更浅）
	for i := 0; i < 5; i++ {
		x1 := grand.N(0, width)
		y1 := grand.N(0, height)
		x2 := grand.N(0, width)
		y2 := grand.N(0, height)

		lineColor := color.RGBA{
			uint8(grand.N(120, 170)),
			uint8(grand.N(120, 170)),
			uint8(grand.N(120, 170)),
			120, // 更透明
		}

		thickness := 1 // 固定1px最干净
		s.drawThickLine(img, x1, y1, x2, y2, lineColor, thickness)
	}

	// 添加干扰点（减少到100个）
	// 添加干扰点（60个，避免太乱）
	for i := 0; i < 60; i++ {
		x := grand.N(0, width)
		y := grand.N(0, height)

		dotColor := color.RGBA{
			uint8(grand.N(120, 200)),
			uint8(grand.N(120, 200)),
			uint8(grand.N(120, 200)),
			180,
		}
		img.Set(x, y, dotColor)
	}

	// 绘制验证码文字（每个字符随机位置和绿色）
	startX := 20
	for i, char := range code {
		// 随机X偏移（-30到+30）
		xOffset := grand.N(-6, 7)
		// 随机Y偏移（-30到+30）
		yOffset := grand.N(-4, 5)
		// 绿色系（深绿到中绿）
		textColor := color.RGBA{
			uint8(grand.N(10, 40)),
			uint8(grand.N(120, 170)),
			uint8(grand.N(10, 40)),
			255,
		}

		x := startX + i*32 + xOffset
		y := 18 + yOffset

		// 先画白色描边（增强可读性）
		stroke := color.RGBA{255, 255, 255, 255}
		s.drawChar(img, x-1, y, string(char), stroke)
		s.drawChar(img, x+1, y, string(char), stroke)
		s.drawChar(img, x, y-1, string(char), stroke)
		s.drawChar(img, x, y+1, string(char), stroke)

		s.drawChar(img, x, y, string(char), textColor)
	}

	// 将图片编码为PNG
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *sPublic) drawThickLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color, thickness int) {
	if thickness <= 1 {
		s.drawLine(img, x1, y1, x2, y2, c)
		return
	}

	half := thickness / 2
	for dx := -half; dx <= half; dx++ {
		for dy := -half; dy <= half; dy++ {
			s.drawLine(img, x1+dx, y1+dy, x2+dx, y2+dy, c)
		}
	}
}

// drawLine 绘制直线
func (s *sPublic) drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	dx := x2 - x1
	dy := y2 - y1
	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}

	if steps == 0 {
		img.Set(x1, y1, c)
		return
	}

	xInc := float64(dx) / float64(steps)
	yInc := float64(dy) / float64(steps)

	x := float64(x1)
	y := float64(y1)

	for i := 0; i <= steps; i++ {
		img.Set(int(x), int(y), c)
		x += xInc
		y += yInc
	}
}

// drawChar 绘制字符（简单实现）
func (s *sPublic) drawChar(img *image.RGBA, x, y int, char string, c color.Color) {
	// 简单的字符绘制（使用点阵）- 支持数字和字母
	patterns := map[string][][]int{
		// 数字
		"0": {{1, 1, 1}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {1, 1, 1}},
		"1": {{0, 1, 0}, {1, 1, 0}, {0, 1, 0}, {0, 1, 0}, {1, 1, 1}},
		"2": {{1, 1, 1}, {0, 0, 1}, {1, 1, 1}, {1, 0, 0}, {1, 1, 1}},
		"3": {{1, 1, 1}, {0, 0, 1}, {1, 1, 1}, {0, 0, 1}, {1, 1, 1}},
		"4": {{1, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 0, 1}, {0, 0, 1}},
		"5": {{1, 1, 1}, {1, 0, 0}, {1, 1, 1}, {0, 0, 1}, {1, 1, 1}},
		"6": {{1, 1, 1}, {1, 0, 0}, {1, 1, 1}, {1, 0, 1}, {1, 1, 1}},
		"7": {{1, 1, 1}, {0, 0, 1}, {0, 0, 1}, {0, 0, 1}, {0, 0, 1}},
		"8": {{1, 1, 1}, {1, 0, 1}, {1, 1, 1}, {1, 0, 1}, {1, 1, 1}},
		"9": {{1, 1, 1}, {1, 0, 1}, {1, 1, 1}, {0, 0, 1}, {1, 1, 1}},
		// 大写字母
		"A": {{0, 1, 0}, {1, 0, 1}, {1, 1, 1}, {1, 0, 1}, {1, 0, 1}},
		"B": {{1, 1, 0}, {1, 0, 1}, {1, 1, 0}, {1, 0, 1}, {1, 1, 0}},
		"C": {{0, 1, 1}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {0, 1, 1}},
		"D": {{1, 1, 0}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {1, 1, 0}},
		"E": {{1, 1, 1}, {1, 0, 0}, {1, 1, 0}, {1, 0, 0}, {1, 1, 1}},
		"F": {{1, 1, 1}, {1, 0, 0}, {1, 1, 0}, {1, 0, 0}, {1, 0, 0}},
		"G": {{0, 1, 1}, {1, 0, 0}, {1, 0, 1}, {1, 0, 1}, {0, 1, 1}},
		"H": {{1, 0, 1}, {1, 0, 1}, {1, 1, 1}, {1, 0, 1}, {1, 0, 1}},
		"J": {{0, 0, 1}, {0, 0, 1}, {0, 0, 1}, {1, 0, 1}, {0, 1, 0}},
		"K": {{1, 0, 1}, {1, 0, 1}, {1, 1, 0}, {1, 0, 1}, {1, 0, 1}},
		"L": {{1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {1, 1, 1}},
		"M": {{1, 0, 1}, {1, 1, 1}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}},
		"N": {{1, 0, 1}, {1, 1, 1}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}},
		"P": {{1, 1, 0}, {1, 0, 1}, {1, 1, 0}, {1, 0, 0}, {1, 0, 0}},
		"Q": {{0, 1, 0}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {0, 1, 1}},
		"R": {{1, 1, 0}, {1, 0, 1}, {1, 1, 0}, {1, 0, 1}, {1, 0, 1}},
		"S": {{0, 1, 1}, {1, 0, 0}, {0, 1, 0}, {0, 0, 1}, {1, 1, 0}},
		"T": {{1, 1, 1}, {0, 1, 0}, {0, 1, 0}, {0, 1, 0}, {0, 1, 0}},
		"U": {{1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {0, 1, 0}},
		"V": {{1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {0, 1, 0}},
		"W": {{1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {1, 1, 1}, {1, 0, 1}},
		"X": {{1, 0, 1}, {1, 0, 1}, {0, 1, 0}, {1, 0, 1}, {1, 0, 1}},
		"Y": {{1, 0, 1}, {1, 0, 1}, {0, 1, 0}, {0, 1, 0}, {0, 1, 0}},
		"Z": {{1, 1, 1}, {0, 0, 1}, {0, 1, 0}, {1, 0, 0}, {1, 1, 1}},
		// 小写字母（简化版）
		"a": {{0, 0, 0}, {0, 1, 1}, {1, 0, 1}, {1, 0, 1}, {0, 1, 1}},
		"b": {{1, 0, 0}, {1, 0, 0}, {1, 1, 0}, {1, 0, 1}, {1, 1, 0}},
		"c": {{0, 0, 0}, {0, 1, 1}, {1, 0, 0}, {1, 0, 0}, {0, 1, 1}},
		"d": {{0, 0, 1}, {0, 0, 1}, {0, 1, 1}, {1, 0, 1}, {0, 1, 1}},
		"e": {{0, 0, 0}, {0, 1, 0}, {1, 1, 1}, {1, 0, 0}, {0, 1, 1}},
		"f": {{0, 1, 1}, {0, 1, 0}, {1, 1, 1}, {0, 1, 0}, {0, 1, 0}},
		"g": {{0, 0, 0}, {0, 1, 1}, {1, 0, 1}, {0, 1, 1}, {1, 1, 0}},
		"h": {{1, 0, 0}, {1, 0, 0}, {1, 1, 0}, {1, 0, 1}, {1, 0, 1}},
		"j": {{0, 0, 1}, {0, 0, 0}, {0, 0, 1}, {0, 0, 1}, {1, 1, 0}},
		"k": {{1, 0, 0}, {1, 0, 1}, {1, 1, 0}, {1, 0, 1}, {1, 0, 1}},
		"m": {{0, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}, {1, 0, 1}},
		"n": {{0, 0, 0}, {1, 1, 0}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}},
		"p": {{0, 0, 0}, {1, 1, 0}, {1, 0, 1}, {1, 1, 0}, {1, 0, 0}},
		"q": {{0, 0, 0}, {0, 1, 1}, {1, 0, 1}, {0, 1, 1}, {0, 0, 1}},
		"r": {{0, 0, 0}, {0, 1, 1}, {0, 1, 0}, {0, 1, 0}, {0, 1, 0}},
		"s": {{0, 0, 0}, {0, 1, 1}, {0, 1, 0}, {0, 0, 1}, {1, 1, 0}},
		"t": {{0, 1, 0}, {1, 1, 1}, {0, 1, 0}, {0, 1, 0}, {0, 0, 1}},
		"u": {{0, 0, 0}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {0, 1, 1}},
		"v": {{0, 0, 0}, {1, 0, 1}, {1, 0, 1}, {1, 0, 1}, {0, 1, 0}},
		"w": {{0, 0, 0}, {1, 0, 1}, {1, 0, 1}, {1, 1, 1}, {1, 0, 1}},
		"x": {{0, 0, 0}, {1, 0, 1}, {0, 1, 0}, {0, 1, 0}, {1, 0, 1}},
		"y": {{0, 0, 0}, {1, 0, 1}, {0, 1, 1}, {0, 0, 1}, {1, 1, 0}},
		"z": {{0, 0, 0}, {1, 1, 1}, {0, 0, 1}, {0, 1, 0}, {1, 1, 1}},
	}

	pattern, ok := patterns[char]
	if !ok {
		return
	}

	// 放大倍数 - 增大到4
	scale := 3

	for row := 0; row < len(pattern); row++ {
		for col := 0; col < len(pattern[row]); col++ {
			if pattern[row][col] == 1 {
				// 绘制放大后的点
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						px := x + col*scale + dx
						py := y + row*scale + dy
						if px < img.Bounds().Max.X && py < img.Bounds().Max.Y {
							if grand.N(0, 100) < 3 { // 3%概率跳过
								continue
							}
							img.Set(px, py, c)
						}
					}
				}
			}
		}
	}
}

// abs 返回绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// RefreshToken 刷新token
func (s *sPublic) RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (res *v1.RefreshTokenRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.RefreshToken", trace.WithAttributes(
		attribute.String("method", "RefreshToken"),
	))
	defer span.End()

	if req.RefreshToken == "" {
		return nil, fmt.Errorf("refresh_token必填")
	}

	// 验证refresh token
	jwtSecret := g.Cfg().MustGet(ctx, "jwt.secret").String()
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret not configured")
	}

	// 解析refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		middleware.LogWithTrace(ctx, "error", "refresh token无效: %v", err)
		return nil, fmt.Errorf("refresh token无效或已过期")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("refresh token格式错误")
	}

	// 验证token类型
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, fmt.Errorf("token类型错误")
	}

	// 检查refresh token是否已被使用（黑名单检查 - 使用jti）
	refreshJti, _ := claims["jti"].(string)
	if refreshJti != "" {
		blacklistKey := fmt.Sprintf("token_blacklist:%s", refreshJti)
		isBlacklisted, err := g.Redis().Exists(ctx, blacklistKey)
		if err == nil && isBlacklisted > 0 {
			middleware.LogWithTrace(ctx, "error", "refresh token已被使用: jti=%s", refreshJti)
			return nil, fmt.Errorf("refresh token已被使用，请使用最新的token")
		}
	}

	// 获取用户信息
	userId, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("refresh token中缺少用户信息")
	}

	siteId, ok := claims["site_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("refresh token中缺少站点信息")
	}

	// 查询用户（确保用户仍然有效）
	var user *entity.User
	err = dao.User.Ctx(ctx).Where("id", uint(userId)).
		Where("site_id", int(siteId)).
		Where("status", 1).
		Scan(&user)

	if err != nil || user == nil {
		middleware.LogWithTrace(ctx, "error", "用户不存在或已禁用: %d", uint(userId))
		return nil, fmt.Errorf("用户不存在或已禁用")
	}

	// 生成新的token和refresh token
	newToken, newRefreshToken, expiresIn, refreshExpiresIn, err := s.generateJWTTokenUser(user)
	if err != nil {
		g.Log().Error(ctx, "生成新token失败:", err)
		return nil, fmt.Errorf("刷新token失败")
	}

	// 将旧的refresh token和access token的jti都加入黑名单
	// refreshJti已在前面定义
	accessTokenJti, _ := claims["access_token_jti"].(string)

	exp, ok := claims["exp"].(float64)
	if ok {
		expTime := time.Unix(int64(exp), 0)
		ttl := time.Until(expTime)
		if ttl > 0 {
			// 将旧refresh token jti加入黑名单
			if refreshJti != "" {
				blacklistKey := fmt.Sprintf("token_blacklist:%s", refreshJti)
				err = g.Redis().SetEX(ctx, blacklistKey, "1", int64(ttl.Seconds()))
				if err != nil {
					middleware.LogWithTrace(ctx, "warn", "添加refresh token jti到黑名单失败: %v", err)
				} else {
					middleware.LogWithTrace(ctx, "info", "旧refresh token jti已加入黑名单: %s, TTL: %v", refreshJti, ttl)
				}
			}

			// 将关联的旧access token jti也加入黑名单
			if accessTokenJti != "" {
				blacklistKey := fmt.Sprintf("token_blacklist:%s", accessTokenJti)
				// access token的TTL使用1小时（access token的有效期）
				accessTTL := int64(3600)
				if ttl.Seconds() < float64(accessTTL) {
					accessTTL = int64(ttl.Seconds())
				}
				err = g.Redis().SetEX(ctx, blacklistKey, "1", accessTTL)
				if err != nil {
					middleware.LogWithTrace(ctx, "warn", "添加access token jti到黑名单失败: %v", err)
				} else {
					middleware.LogWithTrace(ctx, "info", "旧access token jti已加入黑名单: %s, TTL: %d", accessTokenJti, accessTTL)
				}
			}
		}
	}

	middleware.LogWithTrace(ctx, "info", "刷新token成功: user_id=%d, username=%s", user.Id, user.Username)

	res = &v1.RefreshTokenRes{
		Token:            newToken,
		RefreshToken:     newRefreshToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}

	return res, nil
}

// Register 用户注册
func (s *sPublic) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "user.Register", trace.WithAttributes(
		attribute.String("method", "Register"),
		attribute.String("username", req.Username),
	))
	defer span.End()

	// 获取站点ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "site_id_not_found")
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	// 验证图片验证码
	if req.Code == "" {
		return nil, fmt.Errorf("请输入验证码")
	}

	clientIp := util.GetClientIP(ctx)
	siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
	if siteCode == "" {
		siteCode = "default"
	}

	cacheKey := fmt.Sprintf("%s:captcha:%s:%s", siteCode, clientIp, req.Time)
	cachedCode, err := g.Redis().Get(ctx, cacheKey)
	if err != nil || cachedCode.IsNil() {
		return nil, fmt.Errorf("验证码已过期，请重新获取")
	}

	if !strings.EqualFold(cachedCode.String(), req.Code) {
		return nil, fmt.Errorf("验证码错误")
	}

	// 验证成功后删除验证码
	g.Redis().Del(ctx, cacheKey)

	// 验证手机验证码（如果提供）
	// TODO: 实现手机验证码验证逻辑

	// 检查用户名是否已存在
	var existUser *entity.User
	err = dao.User.Ctx(ctx).Where("username", req.Username).
		Where("site_id", siteId).
		Scan(&existUser)
	if err == nil && existUser != nil {
		return nil, fmt.Errorf("用户名已经被注册")
	}

	// 检查手机号是否已存在
	if req.Mobile != "" {
		var existMobile *entity.User
		err = dao.User.Ctx(ctx).Where("mobile", req.Mobile).
			Where("site_id", siteId).
			Scan(&existMobile)
		if err == nil && existMobile != nil {
			return nil, fmt.Errorf("手机账号已经被注册")
		}
	}

	// 加密密码
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		g.Log().Error(ctx, "密码加密失败:", err)
		return nil, fmt.Errorf("注册失败")
	}

	// 加密资金密码（如果提供）
	var hashedPayPassword string
	if req.PayPassword != "" {
		hashedPayPassword, err = s.hashPassword(req.PayPassword)
		if err != nil {
			g.Log().Error(ctx, "资金密码加密失败:", err)
			return nil, fmt.Errorf("注册失败")
		}
	}

	// 获取站点配置
	var siteConfig *entity.SiteConfig
	err = dao.SiteConfig.Ctx(ctx).Where("site_id", siteId).Scan(&siteConfig)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &entity.User{
		SiteId:       int(siteId),
		GradeId:      siteConfig.DefaultGradeId,
		LevelId:      siteConfig.DefaultLevelId,
		AgentId:      siteConfig.DefaultAgentId,
		Username:     req.Username,
		Password:     hashedPassword,
		PayPassword:  hashedPayPassword,
		Realname:     req.Realname,
		Mobile:       req.Mobile,
		Email:        req.Email,
		SafeQuestion: "",
		SafeAnswer:   "",
		RegisterIp:   clientIp,
		RegisterTime: gtime.Now(),
		RegisterUrl:  req.Domain,
		Status:       1,
	}

	// 保存用户
	result, err := dao.User.Ctx(ctx).Insert(user)
	if err != nil {
		g.Log().Error(ctx, "注册失败:", err)
		return nil, fmt.Errorf("注册失败")
	}

	userId, err := result.LastInsertId()
	if err != nil {
		g.Log().Error(ctx, "获取用户ID失败:", err)
		return nil, fmt.Errorf("注册失败")
	}
	user.Id = uint(userId)

	// 更新最后登录信息
	dao.User.Ctx(ctx).Where("id", user.Id).Update(map[string]interface{}{
		"last_login_ip":   clientIp,
		"last_login_time": gtime.Now(),
	})

	// 异步调用balance服务初始化余额
	go func() {
		dbManager := registry.GetDBConfigManagerV2()
		siteConfig, err := dbManager.GetSiteConfig(siteCode)
		if err != nil {
			g.Log().Errorf(ctx, "获取站点配置失败: %v", err)
			return
		}

		currency := siteConfig.Base.Currency
		if currency == "" {
			currency = "CNY"
		}

		msg := &mq.InitBalanceUserMessage{
			SiteID:   int32(siteId),
			UserID:   int32(userId),
			Username: req.Username,
			Currency: currency,
		}

		if err := mq.PublishInitBalanceUser(context.Background(), msg); err != nil {
			g.Log().Errorf(ctx, "发送初始化余额消息失败: %v", err)
		}
	}()

	// 生成JWT Token和Refresh Token
	token, refreshToken, expiresIn, refreshExpiresIn, err := s.generateJWTTokenUser(user)
	if err != nil {
		g.Log().Error(ctx, "生成token失败:", err)
		return nil, fmt.Errorf("注册失败")
	}

	res = &v1.RegisterRes{
		Token:            token,
		RefreshToken:     refreshToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}

	return res, nil
}

// GetSiteSetting 获取站点设置
func (s *sPublic) GetSiteSetting(ctx context.Context, req *v1.GetSiteSettingReq) (res *v1.GetSiteSettingRes, err error) {
	ctx, span := tracing.StartSpan(ctx, "public.GetSiteSetting", trace.WithAttributes(
		attribute.String("method", "GetSiteSetting"),
	))
	defer span.End()

	// 获取站点ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "site_id_not_found")
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	// 查询站点配置
	var siteConfig *entity.SiteConfig
	err = dao.SiteConfig.Ctx(ctx).Where("site_id", siteId).Scan(&siteConfig)
	if err != nil || siteConfig == nil {
		middleware.LogWithTrace(ctx, "error", "查询站点配置失败: %v", err)
		return nil, fmt.Errorf("获取站点配置失败")
	}

	// 查询站点信息
	var site *entity.Site
	err = dao.Site.Ctx(ctx).Where("id", siteId).Scan(&site)
	if err != nil || site == nil {
		middleware.LogWithTrace(ctx, "error", "查询站点信息失败: %v", err)
		return nil, fmt.Errorf("获取站点信息失败")
	}

	// 默认关闭理由
	closeReason := "您好，系统正在维护中，估计需要1个小时左右，给您造成不便，敬请谅解！"
	if siteConfig.CloseReason != "" {
		closeReason = siteConfig.CloseReason
	}

	res = &v1.GetSiteSettingRes{
		AgentUrl:       siteConfig.UrlAgentPc,
		SiteName:       site.Name,
		MobileUrl:      siteConfig.UrlMobile,
		SwitchRegister: int32(siteConfig.SwitchRegister),
		SwitchGrade:    int32(siteConfig.SwitchGrade),
		ServiceUrl:     siteConfig.UrlService,
		Close:          int32(siteConfig.IsClose),
		MinWithdraw:    fmt.Sprintf("%.2f", float64(siteConfig.MinWithdraw)),
		MaxWithdraw:    fmt.Sprintf("%.2f", float64(siteConfig.MaxWithdraw)),
		MobileLogo:     siteConfig.MobileLogo,
		CloseReason:    closeReason,
	}

	return res, nil
}
