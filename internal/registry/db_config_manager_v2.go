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
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/metadata"
)

// SiteConfig 站点配置结构（匹配新的KV格式）
type SiteConfig struct {
	Meta   MetaConfig   `json:"Meta"`
	Base   BaseConfig   `json:"Base"`
	Domain DomainConfig `json:"Domain"`
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
		g.Log().Infof(m.ctx, "App服务V2站点配置已更新（仅业务配置）: site=%s, version=%d", siteCode, config.Meta.Version)
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

	// 动态设置数据库配置（这里会预建连接池）
	gdb.AddConfigNode(groupName, gdb.ConfigNode{
		Link: config.Base.DbLinkSite,
	})

	g.Log().Infof(m.ctx, "已注册数据库连接池: %s -> %s", groupName, config.Base.DbLinkSite)
	return nil
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
