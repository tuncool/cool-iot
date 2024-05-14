package models

import (
	"cool-iot/pkg/crypt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type AuthUser struct {
	Id           int32     `gorm:"column:id;type:int;primaryKey;autoIncrement" json:"-"`
	Uid          int64     `gorm:"column:uid;type:bigint(20);not null;unique;index" json:"uid,omitempty"`
	Level        int32     `gorm:"column:level;type:int;not null;default:99" json:"level,omitempty"`
	Username     string    `gorm:"column:username;type:varchar(30);not null;default:guest" json:"username"`
	Password     string    `gorm:"column:password;type:varchar(128);not null;default:12345" json:"password,omitempty"`
	Emile        string    `gorm:"column:emil;type:varchar(30);not null;default:admin@tunps.com" json:"emile,omitempty"`
	OTAKeys      string    `gorm:"column:ota_keys;type:varchar(255);not null;default:tunpxs" json:"ota_keys,omitempty"`
	OTAEnable    bool      `gorm:"column:ota_keys;not null;default:false" json:"ota_enable,omitempty"`
	LastLoginIp  string    `gorm:"column:last_login_ip;type:varchar(20);not null;default:127.0.0.1" json:"last_login_ip,omitempty"`
	CreateAt     time.Time `gorm:"column:create_at;autoCreateTime" json:"create_at"`
	LastLoginAt  time.Time `gorm:"column:last_login_at" json:"last_login_at"`
	ExpirationAt int64     `gorm:"column:expiration_at;type:int;not null;default:0" json:"expiration_at"`
	Valid        bool      `gorm:"column:valid;not null;default:false" json:"valid"`
}

func (*AuthUser) TableName() string {
	return "auth_user"
}

func (m *AuthUser) BeforeCreate(tx *gorm.DB) (err error) {
	if m.Uid == 0 {
		m.Uid = crypt.SnowID(rand.Int63n(1024))
	}
	m.CreateAt = time.Now()
	m.LastLoginAt = time.Now()
	return
}

func (m *AuthUser) BeforeUpdate(tx *gorm.DB) (err error) {
	m.LastLoginAt = time.Now()
	return
}
