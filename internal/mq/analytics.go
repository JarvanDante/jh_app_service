package mq

// SyncUserToClickHouseMessage 同步用户到 ClickHouse
type SyncUserToClickHouseMessage struct {
	EventID      uint64 `json:"event_id"`
	SiteID       uint32 `json:"site_id"`
	UserID       uint64 `json:"user_id"`
	InviterID    uint64 `json:"inviter_id"`
	Username     string `json:"username"`
	Status       uint8  `json:"status"`
	RegisterIP   string `json:"register_ip"`
	RegisterTime string `json:"register_time"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	PayTimes     uint32 `json:"pay_times"`
}
