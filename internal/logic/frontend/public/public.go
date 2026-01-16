package public

import (
	"context"
	"fmt"
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
	"jh_app_service/internal/service/frontend"
	"jh_app_service/internal/tracing"
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
	// TODO: 验证验证码
	// isValid := checkCaptcha(req.Code, req.Time)
	// if !isValid {
	//     return nil, gerror.New("请输入正确的验证码")
	// }

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

	// TODO: 以下操作需要调用其他微服务或异步处理
	// 1. 添加会员余额 - 调用 balance_service
	// 2. 代理注册奖金 - 调用 balance_service
	// 3. 风险预警 - 异步处理
	// 4. 自定义资源绑定 - 异步处理

	// 生成JWT Token
	token, err := s.generateJWTTokenUser(user)
	if err != nil {
		g.Log().Error(ctx, "生成token失败:", err)
		return nil, fmt.Errorf("注册失败")
	}

	res = &v1.QuickRegisterRes{
		Token:    token,
		Username: username,
		Password: password,
	}

	return res, nil
}

// generateJWTTokenUser 生成JWT token
func (s *sPublic) generateJWTTokenUser(user *entity.User) (string, error) {
	// 从配置文件获取JWT密钥
	jwtSecret := g.Cfg().MustGet(context.Background(), "jwt.secret").String()
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		//"user_id":  admin.Id, // 使用 user_id 字段名，与 Gateway 的 Claims 结构体匹配
		"user_id":  user.Id, // 后台使用 user_id =0
		"admin_id": 0,       // 保留 admin_id 用于兼容
		"username": user.Username,
		"site_id":  user.SiteId,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
		"iat":      time.Now().Unix(),
	})

	// 签名token
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
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

	// 获取站点ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "site_id_not_found")
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	// TODO: 验证验证码（如果需要）
	// if req.Code != "" {
	//     isValid := checkCaptcha(req.Code, req.Time)
	//     if !isValid {
	//         return nil, fmt.Errorf("验证码错误")
	//     }
	// }

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
	_, err = dao.User.Ctx(ctx).Where("id", user.Id).Update(map[string]interface{}{
		"last_login_ip":   "127.0.0.1", // TODO: 获取真实IP
		"last_login_time": gtime.Now(),
	})

	// TODO: 记录登录日志
	// TODO: 清空登录失败次数

	// 生成JWT Token
	token, err := s.generateJWTTokenUser(user)
	if err != nil {
		g.Log().Error(ctx, "生成token失败:", err)
		return nil, fmt.Errorf("登录失败")
	}

	res = &v1.LoginRes{
		Token: token,
	}

	return res, nil
}
