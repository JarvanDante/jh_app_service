package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/metadata"
)

// SiteConfig 站点配置结构（匹配新的KV格式）
type SiteConfig struct {
	Meta     MetaConfig       `json:"Meta"`
	Base     BaseConfig       `json:"Base"`
	Domain   DomainConfig     `json:"Domain"`
	Activity []ActivityConfig `json:"Activity"`
}

// MetaConfig 元数据配置
type MetaConfig struct {
	Version     int64  `json:"version"`
	PublishedAt string `json:"published_at"`
	AdminId     int    `json:"admin_id"`
}

// BaseConfig 基础配置
type BaseConfig struct {
	SiteId        int    `json:"site_id"`
	Code          string `json:"code"`
	TimeZone      string `json:"timezone"`
	Currency      string `json:"currency"`
	DbLinkSite    string `json:"db_link_site"`
	DbLinkBalance string `json:"db_link_balance"`
	DbLinkGame    string `json:"db_link_game"`
	Status        int    `json:"status"`
}

// DomainConfig 域名配置
type DomainConfig struct {
	API      []string `json:"api"`
	H5       []string `json:"h5"`
	Admin    []string `json:"admin"`
	Callback []string `json:"callback"`
}

// ActivityConfig 活动配置
type ActivityConfig struct {
	ActivityId   int    `json:"activity_id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	ActivityType string `json:"activity_type"`
	Uri          string `json:"uri"`
	Table        string `json:"table"`
	FixedParams  string `json:"fixed_params"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Status       int    `json:"status"`
	Sort         int    `json:"sort"`
}

// DBConfigManagerV2 数据库配置管理器V2 - 无侵入版本
type DBConfigManagerV2 struct {
	consul      *api.Client
	siteConfigs map[string]*SiteConfig // key: site_code, value: config
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

var (
	dbManagerV2 *DBConfigManagerV2
	onceV2      sync.Once
)

// GetDBConfigManagerV2 获取数据库配置管理器V2单例
func GetDBConfigManagerV2() *DBConfigManagerV2 {
	onceV2.Do(func() {
		dbManagerV2 = &DBConfigManagerV2{
			siteConfigs: make(map[string]*SiteConfig),
		}
	})
	return dbManagerV2
}

// Init 初始化数据库配置管理器
func (m *DBConfigManagerV2) Init(consulClient *api.Client) error {
	m.consul = consulClient
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// 初始加载所有站点配置
	if err := m.loadAllSiteConfigs(); err != nil {
		g.Log().Errorf(m.ctx, "初始加载站点配置失败: %v", err)
		return err
	}

	// 启动配置监听
	go m.watchSiteConfigs()

	g.Log().Info(m.ctx, "App服务数据库配置管理器V2初始化完成")
	return nil
}

// Stop 停止配置管理器
func (m *DBConfigManagerV2) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	g.Log().Info(m.ctx, "App服务数据库配置管理器V2已停止")
}

// loadAllSiteConfigs 加载所有站点配置
func (m *DBConfigManagerV2) loadAllSiteConfigs() error {
	kv := m.consul.KV()

	// 获取所有 sites/ 前缀的配置
	pairs, _, err := kv.List("sites/", nil)
	if err != nil {
		return fmt.Errorf("获取站点配置列表失败: %v", err)
	}

	updatedCount := 0
	for _, pair := range pairs {
		// 解析 key: sites/{site_code}
		parts := strings.Split(pair.Key, "/")
		if len(parts) != 2 {
			continue
		}

		updated, err := m.updateSiteConfig(pair.Key, pair.Value)
		if err != nil {
			g.Log().Errorf(m.ctx, "更新站点配置失败: key=%s, error=%v", pair.Key, err)
		} else if updated {
			updatedCount++
		}
	}

	// 只在有配置更新时输出日志
	if updatedCount > 0 {
		g.Log().Infof(m.ctx, "App服务V2更新了 %d 个站点配置，当前共 %d 个站点", updatedCount, len(m.siteConfigs))
	}
	return nil
}

// watchSiteConfigs 监听站点配置变化
func (m *DBConfigManagerV2) watchSiteConfigs() {
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			// 每30秒检查一次配置变化
			time.Sleep(30 * time.Second)
			if err := m.loadAllSiteConfigs(); err != nil {
				g.Log().Errorf(m.ctx, "重新加载站点配置失败: %v", err)
			}
		}
	}
}

// updateSiteConfig 更新站点配置，返回是否有更新
func (m *DBConfigManagerV2) updateSiteConfig(key string, value []byte) (bool, error) {
	var config SiteConfig
	if err := json.Unmarshal(value, &config); err != nil {
		return false, fmt.Errorf("解析站点配置失败: %v", err)
	}

	siteCode := config.Base.Code
	if siteCode == "" {
		return false, fmt.Errorf("站点代码为空")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查配置是否有变化
	oldConfig, exists := m.siteConfigs[siteCode]
	if exists && oldConfig.Meta.Version >= config.Meta.Version {
		return false, nil // 配置没有变化或版本更旧
	}

	// 检查数据库连接是否变化
	dbChanged := false
	if !exists {
		// 新站点，需要注册连接池
		dbChanged = true
	} else {
		// 检查三个数据库连接是否有任何变化
		if oldConfig.Base.DbLinkSite != config.Base.DbLinkSite ||
			oldConfig.Base.DbLinkBalance != config.Base.DbLinkBalance ||
			oldConfig.Base.DbLinkGame != config.Base.DbLinkGame {
			dbChanged = true
			g.Log().Infof(m.ctx, "站点 %s 数据库连接配置发生变化，将重新注册连接池", siteCode)
		}
	}

	// 更新配置到内存
	m.siteConfigs[siteCode] = &config

	// 只在数据库连接变化时才重新注册连接池
	if dbChanged {
		if err := m.registerDBGroup(siteCode, &config); err != nil {
			g.Log().Errorf(m.ctx, "注册数据库配置组失败: site=%s, error=%v", siteCode, err)
			return false, err
		}
		g.Log().Infof(m.ctx, "App服务V2站点配置已更新（含DB连接池）: site=%s, version=%d", siteCode, config.Meta.Version)
	} else {
		g.Log().Debugf(m.ctx, "App服务V2站点配置已更新（仅业务配置）: site=%s, version=%d", siteCode, config.Meta.Version)
	}

	// 同步活动配置到站点数据库
	if err := m.syncActivityConfig(siteCode, &config); err != nil {
		g.Log().Warningf(m.ctx, "同步站点活动配置失败: site=%s, error=%v", siteCode, err)
	}

	return true, nil
}

// registerDBGroup 动态注册数据库配置组（预建连接池）
func (m *DBConfigManagerV2) registerDBGroup(siteCode string, config *SiteConfig) error {
	if config.Base.DbLinkSite == "" {
		g.Log().Warningf(m.ctx, "站点 %s 没有配置Site数据库连接", siteCode)
		return nil
	}

	// 创建数据库配置组名称
	groupName := fmt.Sprintf("site_%s", siteCode)

	// 解析连接字符串
	// 格式: mysql:root:654321@tcp(172.19.0.4:3306)/jh_site?parseTime=true&loc=Local
	link := config.Base.DbLinkSite

	// 移除 "mysql:" 前缀
	if len(link) > 6 && link[:6] == "mysql:" {
		link = link[6:] // 移除 "mysql:"
	}

	configNode := gdb.ConfigNode{
		Type: "mysql",
		Link: link, // 使用移除了type前缀的link
	}

	gdb.AddConfigNode(groupName, configNode)

	// 验证配置并测试连接
	registeredConfig := gdb.GetConfig(groupName)
	if registeredConfig != nil && len(registeredConfig) > 0 {
		// 尝试实际连接测试
		testDB := g.DB(groupName)
		if testDB != nil {
			_, err := testDB.GetValue(m.ctx, "SELECT 1")
			if err != nil {
				g.Log().Errorf(m.ctx, "数据库连接测试失败: %s, error: %v", groupName, err)
				return fmt.Errorf("数据库连接测试失败: %v", err)
			}
			g.Log().Debugf(m.ctx, "数据库连接池注册成功: %s", groupName)
		}
	} else {
		g.Log().Errorf(m.ctx, "配置注册失败: %s", groupName)
		return fmt.Errorf("配置注册失败")
	}

	return nil
}

// syncActivityConfig 同步活动配置到站点数据库
func (m *DBConfigManagerV2) syncActivityConfig(siteCode string, config *SiteConfig) error {
	groupName := fmt.Sprintf("site_%s", siteCode)
	db := g.DB(groupName)
	if db == nil {
		g.Log().Warningf(m.ctx, "站点 %s 数据库连接不存在，跳过活动配置同步", siteCode)
		return nil
	}

	// 检查数据库表
	tables, err := db.Tables(m.ctx)
	if err != nil {
		return fmt.Errorf("查询数据库表失败: %v", err)
	}

	tableSet := make(map[string]bool, len(tables))
	hasTable := false
	for _, table := range tables {
		tableSet[table] = true
		if table == "site_activity" {
			hasTable = true
		}
	}
	if !hasTable {
		g.Log().Warningf(m.ctx, "站点 %s 数据库中不存在 site_activity 表，跳过活动配置同步", siteCode)
		return nil
	}

	fields, err := db.TableFields(m.ctx, "site_activity")
	if err != nil {
		return fmt.Errorf("查询 site_activity 表字段失败: %v", err)
	}
	tableFieldsCache := map[string]map[string]*gdb.TableField{
		"site_activity": fields,
	}

	hasField := func(name string) bool {
		_, ok := fields[name]
		return ok
	}

	// 站点过滤字段
	siteFilter := g.Map{}
	if hasField("site_id") {
		siteFilter["site_id"] = config.Base.SiteId
	}
	if hasField("site_code") {
		siteFilter["site_code"] = config.Base.Code
	}
	if len(siteFilter) == 0 {
		g.Log().Warningf(m.ctx, "site_activity 表缺少 site_id/site_code 字段，跳过同步: site=%s", siteCode)
		return nil
	}

	// 主键字段（优先 activity_id）
	keyField := ""
	if hasField("activity_id") {
		keyField = "activity_id"
	} else if hasField("code") {
		keyField = "code"
	} else {
		g.Log().Warningf(m.ctx, "site_activity 表缺少 activity_id/code 字段，跳过同步: site=%s", siteCode)
		return nil
	}

	// 查询现有活动
	type keyRow struct {
		ActivityId int    `json:"activity_id"`
		Code       string `json:"code"`
	}
	var existing []keyRow
	query := db.Model("site_activity").Fields(keyField).Where(siteFilter)
	if err := query.Scan(&existing); err != nil {
		return fmt.Errorf("查询站点活动失败: %v", err)
	}

	existingMap := make(map[string]bool)
	for _, row := range existing {
		if keyField == "activity_id" {
			existingMap[fmt.Sprintf("%d", row.ActivityId)] = true
		} else {
			existingMap[row.Code] = true
		}
	}

	g.Log().Infof(m.ctx, "开始同步站点 %s 活动配置，共 %d 个活动", siteCode, len(config.Activity))

	return db.Transaction(m.ctx, func(ctx context.Context, tx gdb.TX) error {
		configKeys := make(map[string]bool)

		buildData := func(activity ActivityConfig, insert bool) g.Map {
			data := g.Map{}
			if hasField("site_id") {
				data["site_id"] = config.Base.SiteId
			}
			if hasField("site_code") {
				data["site_code"] = config.Base.Code
			}
			if hasField("activity_id") {
				data["activity_id"] = activity.ActivityId
			}
			if hasField("name") {
				data["name"] = activity.Name
			}
			if hasField("code") {
				data["code"] = activity.Code
			}
			if hasField("activity_type") {
				data["activity_type"] = activity.ActivityType
			}
			if hasField("uri") {
				data["uri"] = activity.Uri
			}
			if hasField("table") {
				data["table"] = activity.Table
			}
			if hasField("fixed_params") {
				data["fixed_params"] = activity.FixedParams
			}
			if hasField("start_time") && activity.StartTime != "" {
				if t, err := gtime.StrToTime(activity.StartTime); err == nil {
					data["start_time"] = t
				} else {
					data["start_time"] = activity.StartTime
				}
			}
			if hasField("end_time") && activity.EndTime != "" {
				if t, err := gtime.StrToTime(activity.EndTime); err == nil {
					data["end_time"] = t
				} else {
					data["end_time"] = activity.EndTime
				}
			}
			if hasField("status") {
				data["status"] = activity.Status
			}
			if hasField("sort") {
				data["sort"] = activity.Sort
			}
			if insert && hasField("created_at") {
				data["created_at"] = gtime.Now()
			}
			if hasField("updated_at") {
				data["updated_at"] = gtime.Now()
			}
			return data
		}

		protectKVFields := func(ctx context.Context, tx gdb.TX, key string, data g.Map) g.Map {
			if len(data) == 0 {
				return data
			}
			// 如果没有这些字段，直接返回
			if !hasField("description") && !hasField("start_time") && !hasField("end_time") &&
				!hasField("fixed_params") && !hasField("status") && !hasField("sort") {
				return data
			}

			type existingFields struct {
				Description string      `json:"description"`
				StartTime   *gtime.Time `json:"start_time"`
				EndTime     *gtime.Time `json:"end_time"`
				FixedParams string      `json:"fixed_params"`
				Status      *int        `json:"status"`
				Sort        *int        `json:"sort"`
			}

			var existing existingFields
			if err := tx.Model("site_activity").
				Fields("description,start_time,end_time,fixed_params,status,sort").
				Where(siteFilter).
				Where(keyField, key).
				Scan(&existing); err != nil {
				g.Log().Warningf(ctx, "查询活动已有字段失败: key=%s, error=%v", key, err)
				return data
			}

			if hasField("description") && strings.TrimSpace(existing.Description) != "" {
				delete(data, "description")
			}
			if hasField("start_time") && existing.StartTime != nil && !existing.StartTime.IsZero() {
				delete(data, "start_time")
			}
			if hasField("end_time") && existing.EndTime != nil && !existing.EndTime.IsZero() {
				delete(data, "end_time")
			}
			if hasField("fixed_params") && strings.TrimSpace(existing.FixedParams) != "" {
				delete(data, "fixed_params")
			}
			if hasField("status") && existing.Status != nil {
				delete(data, "status")
			}
			if hasField("sort") && existing.Sort != nil {
				delete(data, "sort")
			}

			return data
		}

		syncFixedParams := func(activity ActivityConfig) {
			if activity.Table == "" {
				return
			}
			if activity.FixedParams == "" {
				return
			}
			if !tableSet[activity.Table] {
				g.Log().Warningf(ctx, "活动参数表不存在，跳过同步: site=%s, activity=%d, table=%s",
					siteCode, activity.ActivityId, activity.Table)
				return
			}

			activityFields, ok := tableFieldsCache[activity.Table]
			if !ok {
				f, err := db.TableFields(ctx, activity.Table)
				if err != nil {
					g.Log().Errorf(ctx, "查询活动参数表字段失败: table=%s, error=%v", activity.Table, err)
					return
				}
				tableFieldsCache[activity.Table] = f
				activityFields = f
			}

			var params map[string]interface{}
			if err := json.Unmarshal([]byte(activity.FixedParams), &params); err != nil {
				g.Log().Errorf(ctx, "解析 fixed_params 失败: site=%s, activity=%d, error=%v",
					siteCode, activity.ActivityId, err)
				return
			}

			data := g.Map{}
			for key, value := range params {
				if _, ok := activityFields[key]; ok {
					data[key] = value
				}
			}
			if len(data) == 0 {
				return
			}

			where := g.Map{}
			if _, ok := activityFields["site_id"]; ok {
				where["site_id"] = config.Base.SiteId
			}
			if _, ok := activityFields["site_code"]; ok {
				where["site_code"] = config.Base.Code
			}
			if _, ok := activityFields["activity_id"]; ok {
				where["activity_id"] = activity.ActivityId
			}
			if len(where) == 0 {
				g.Log().Warningf(ctx, "活动参数表缺少定位字段，跳过同步: table=%s, activity=%d",
					activity.Table, activity.ActivityId)
				return
			}

			if _, ok := activityFields["updated_at"]; ok {
				data["updated_at"] = gtime.Now()
			}

			count, err := tx.Model(activity.Table).Where(where).Count()
			if err != nil {
				g.Log().Errorf(ctx, "检查活动参数是否存在失败: table=%s, activity=%d, error=%v",
					activity.Table, activity.ActivityId, err)
				return
			}

			if count > 0 {
				if _, err := tx.Model(activity.Table).Where(where).Data(data).Update(); err != nil {
					g.Log().Errorf(ctx, "更新活动参数失败: table=%s, activity=%d, error=%v",
						activity.Table, activity.ActivityId, err)
				}
				return
			}

			if _, ok := activityFields["site_id"]; ok {
				data["site_id"] = config.Base.SiteId
			}
			if _, ok := activityFields["site_code"]; ok {
				data["site_code"] = config.Base.Code
			}
			if _, ok := activityFields["activity_id"]; ok {
				data["activity_id"] = activity.ActivityId
			}
			if _, ok := activityFields["created_at"]; ok {
				data["created_at"] = gtime.Now()
			}
			if _, ok := activityFields["updated_at"]; ok {
				data["updated_at"] = gtime.Now()
			}

			if _, err := tx.Model(activity.Table).Data(data).Insert(); err != nil {
				g.Log().Errorf(ctx, "插入活动参数失败: table=%s, activity=%d, error=%v",
					activity.Table, activity.ActivityId, err)
			}
		}

		for _, activity := range config.Activity {
			key := ""
			if keyField == "activity_id" {
				key = fmt.Sprintf("%d", activity.ActivityId)
			} else {
				key = activity.Code
			}
			configKeys[key] = true

			count, err := tx.Model("site_activity").
				Where(siteFilter).
				Where(keyField, key).
				Count()
			if err != nil {
				g.Log().Errorf(ctx, "检查活动是否存在失败: key=%s, error=%v", key, err)
				continue
			}

			if count > 0 {
				updateData := buildData(activity, false)
				updateData = protectKVFields(ctx, tx, key, updateData)
				_, err = tx.Model("site_activity").
					Where(siteFilter).
					Where(keyField, key).
					Data(updateData).
					Update()
				if err != nil {
					g.Log().Errorf(ctx, "更新活动配置失败: key=%s, error=%v", key, err)
				}
			} else {
				_, err = tx.Model("site_activity").
					Data(buildData(activity, true)).
					Insert()
				if err != nil {
					g.Log().Errorf(ctx, "插入活动配置失败: key=%s, error=%v", key, err)
				}
			}

			// 同步 fixed_params 到活动参数表
			syncFixedParams(activity)
		}

		// 删除配置中不存在的活动
		for key := range existingMap {
			if !configKeys[key] {
				_, err := tx.Model("site_activity").
					Where(siteFilter).
					Where(keyField, key).
					Delete()
				if err != nil {
					g.Log().Errorf(ctx, "删除活动配置失败: key=%s, error=%v", key, err)
				}
			}
		}

		return nil
	})
}

// GetDynamicDB 根据上下文获取动态数据库连接
func (m *DBConfigManagerV2) GetDynamicDB(ctx context.Context) gdb.DB {
	// 获取站点代码
	siteCode := m.getSiteCodeFromContext(ctx)
	if siteCode == "" {
		g.Log().Debugf(ctx, "使用默认数据库连接")
		return g.DB() // 使用默认数据库
	}

	// 尝试获取站点专用数据库
	groupName := fmt.Sprintf("site_%s", siteCode)

	// 检查配置是否存在
	if gdb.GetConfig(groupName) != nil {
		g.Log().Debugf(ctx, "使用站点数据库: %s", groupName)
		return g.DB(groupName)
	}

	g.Log().Warningf(ctx, "站点数据库配置不存在，使用默认数据库: %s", groupName)
	return g.DB() // 回退到默认数据库
}

// getSiteCodeFromContext 从上下文获取站点代码
func (m *DBConfigManagerV2) getSiteCodeFromContext(ctx context.Context) string {
	// 首先尝试从上下文中获取site_code
	if siteCode := ctx.Value("site_code"); siteCode != nil {
		if siteCodeStr, ok := siteCode.(string); ok {
			return siteCodeStr
		}
	}

	// 尝试从gRPC metadata获取site_code
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if siteCodes := md.Get("site_code"); len(siteCodes) > 0 {
			if siteCode := strings.TrimSpace(siteCodes[0]); siteCode != "" {
				return siteCode
			}
		}

		// 如果没有site_code，尝试从site_id查找
		if siteIDs := md.Get("site_id"); len(siteIDs) > 0 {
			if siteIDStr := strings.TrimSpace(siteIDs[0]); siteIDStr != "" {
				if siteID, err := strconv.ParseInt(siteIDStr, 10, 64); err == nil && siteID > 0 {
					// 根据site_id查找对应的site_code
					m.mutex.RLock()
					defer m.mutex.RUnlock()
					for code, config := range m.siteConfigs {
						if int64(config.Base.SiteId) == siteID {
							g.Log().Debugf(ctx, "site_id=%d 对应 site_code=%s", siteID, code)
							return code
						}
					}
				}
			}
		}
	}

	return ""
}

// GetSiteConfig 获取站点配置
func (m *DBConfigManagerV2) GetSiteConfig(siteCode string) (*SiteConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config, exists := m.siteConfigs[siteCode]
	if !exists {
		return nil, fmt.Errorf("站点配置不存在: site=%s", siteCode)
	}

	return config, nil
}

// GetAllSiteCodes 获取所有站点代码
func (m *DBConfigManagerV2) GetAllSiteCodes() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	codes := make([]string, 0, len(m.siteConfigs))
	for code := range m.siteConfigs {
		codes = append(codes, code)
	}

	return codes
}

// GetSiteCodeBySiteID 根据站点ID获取站点代码
func (m *DBConfigManagerV2) GetSiteCodeBySiteID(siteID int64) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for code, config := range m.siteConfigs {
		if int64(config.Base.SiteId) == siteID {
			return code
		}
	}

	return ""
}
