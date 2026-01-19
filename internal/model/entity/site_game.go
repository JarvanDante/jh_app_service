// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SiteGame is the golang structure for table site_game.
type SiteGame struct {
	Id        uint        `json:"id"        orm:"id"         description:""`
	SiteId    int         `json:"siteId"    orm:"site_id"    description:"站点ID"`
	GameId    int         `json:"gameId"    orm:"game_id"    description:"总站里游戏ID"`
	Type      int         `json:"type"      orm:"type"       description:"游戏类型。1=体育；2=彩票；3=真人；4=电子游戏"`
	Name      string      `json:"name"      orm:"name"       description:"游戏名称"`
	Platform  string      `json:"platform"  orm:"platform"   description:"游戏平台"`
	ImageCode string      `json:"imageCode" orm:"image_code" description:"游戏图片目录"`
	DbName    string      `json:"dbName"    orm:"db_name"    description:"游戏对应数据表"`
	ApiId     string      `json:"apiId"     orm:"api_id"     description:"API_ID"`
	ApiKey    string      `json:"apiKey"    orm:"api_key"    description:"API_KEY"`
	ApiUrl    string      `json:"apiUrl"    orm:"api_url"    description:"API_URL"`
	Status    int         `json:"status"    orm:"status"     description:"游戏是否打开或者或者关闭。1=打开；0=关闭"`
	Sort      int         `json:"sort"      orm:"sort"       description:""`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
}
