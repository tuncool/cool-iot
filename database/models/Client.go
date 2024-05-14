package models

import (
	"gorm.io/gorm"
	"time"
)

type Client struct {
	Id         int64     `gorm:"column:id;type:integer;auto_increment;not null;primaryKey;" json:"id"`
	Uid        int64     `gorm:"column:uid;type:integer;not null" json:"uid"`           // 所属分组ID
	GroupId    int64     `gorm:"column:group_id;type:integer;not null" json:"group_id"` // 所属分组ID
	Name       string    `gorm:"column:name;type:varchar(40);not null" json:"name"`     // 名称
	DevType    int32     `gorm:"column:dev_type;type:int(11);not null;default:0" json:"dev_type"`
	AccessId   string    `gorm:"column:access_id;type:varchar(64);not null" json:"access_id"`       // 认证ID
	AccessKey  string    `gorm:"column:access_key;type:varchar(128);not null" json:"access_key"`    // 认证密钥
	RemoteAddr string    `gorm:"column:remote_addr;type:varchar(20);: " json:"remote_addr"`         // 客户端地址
	Remark     string    `gorm:"column:remark;type:varchar(100);" json:"remark"`                    // 备注
	Valid      bool      `gorm:"column:valid;not null;default:true" json:"valid"`                   // 是否启用
	Connected  bool      `gorm:"column:connected;not null;default:true" json:"connected"`           // 是否已经连接
	Version    string    `gorm:"column:version;type:varchar(20);not null;default:0" json:"version"` // 客户端的软件版本
	CreateAt   time.Time `gorm:"column:create_at;autoCreateTime"`                                   // 创建时间
	ActiveAt   time.Time `gorm:"column:active_at;autoUpdateTime" json:"active_at"`                  // 活跃时间，上一次连接时间，断开连接时间
	BlackId    uint32    `gorm:"column:black_id;type:integer;default:0;not null" json:"blackId"`    // 黑名单分组ID
}

func (*Client) TableName() string {
	return "client"
}

func (m *Client) BeforeCreate(tx *gorm.DB) (err error) {
	m.CreateAt = time.Now()
	m.ActiveAt = time.Now()
	return
}

func (m *Client) BeforeUpdate(tx *gorm.DB) (err error) {
	m.ActiveAt = time.Now()
	return
}
