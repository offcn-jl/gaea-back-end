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
	// 腾讯云相关配置
	TencentCloudAPISecretID  string // 令牌
	TencentCloudAPISecretKey string // 密钥
	TencentCloudSmsSdkAppId  string // 短信应用 ID
	// 内部服务相关配置
	ServicesAccessToken string // 接口访问令牌
}
