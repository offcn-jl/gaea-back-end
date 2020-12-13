/*
   @Time : 2020/10/31 3:20 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : system
   @Software: GoLand
   @Description: 结构体 系统
*/

package structs

import (
	"github.com/jinzhu/gorm"
	"time"
)

// SystemConfig 系统配置表
type SystemConfig struct {
	gorm.Model
	DisableDebug bool // 关闭调试模式, 由于 bool 类型的默认初始值为 false 为了在没有初始化成功的情况下默认开启调试, 所以使用 禁用调试 替代 开启调试
	// 跨域检查规则
	CORSRuleServices string // Service 内部服务路由组
	CORSRuleManages  string // Manages 管理平台路由组
	CORSRuleEvents   string // Events 活动 ( 外部服务 ) 路由组
	// 中公教育内部平台相关配置
	OffcnSmsURL      string // 短信平台 接口地址
	OffcnSmsUserName string // 短信平台 用户名
	OffcnSmsPassword string // 短信平台 密码
	OffcnSmsTjCode   string // 短信平台 发送方识别码
	OffcnMisURL      string // 口令码平台 接口地址
	OffcnMisAppID    string // 口令码平台 应用 ID
	OffcnMisToken    string // 口令码平台 令牌
	OffcnMisCode     string // 口令码平台 签名密钥
	// 腾讯云相关配置
	TencentCloudAPISecretID  string // 令牌
	TencentCloudAPISecretKey string // 密钥
	TencentCloudSmsSdkAppId  string // 短信应用 ID
	// 内部服务相关配置
	ServicesAccessToken string // 接口访问令牌
	// RSA 签名密钥
	RSAPublicKey  string `gorm:"type:varchar(1000);"` // RSA 公钥
	RSAPrivateKey string `gorm:"type:varchar(4000);"` // RSA 私钥
}

// SystemUser 系统用户表
type SystemUser struct {
	gorm.Model
	CreatedUserID uint   `gorm:"not null"`                    // 创建用户 ID
	UpdatedUserID uint   `gorm:"not null"`                    // 最终修改用户 ID
	RoleID        uint   `gorm:"not null"`                    // 角色 ID
	Username      string `gorm:"not null"`                    // 用户名
	Password      string `gorm:"type:varchar(1000);not null"` // 密码
	Name          string `gorm:"not null"`                    // 姓名
}

// SystemUserLoginFailLog 系统用户登陆失败日志
type SystemUserLoginFailLog struct {
	gorm.Model
	UserID   uint   `gorm:"not null"` // 用户 ID
	Password string `gorm:"not null"` // 使用的密码
	SourceIp string `gorm:"not null"` // 访问 IP
}

// SystemSession 系统会话表
type SystemSession struct {
	gorm.Model
	UUID          string    `gorm:"not null"` // 会话 ID, 使用随机生成的 UUID 作为会话识别代码, 而不使用自增的 ID 作为会话识别代码, 原因是自增的代码可以很容易的被猜测到, 存在很大的被仿冒的风险
	UserID        uint      `gorm:"not null"` // 用户 ID
	MisToken      string    `gorm:"not null"` // Mis 令牌
	LastRequestAt time.Time `gorm:"not null"` // 最后一次操作时间
	LastSourceIP  string    `gorm:"not null"` // 最后一次操作 IP
}
