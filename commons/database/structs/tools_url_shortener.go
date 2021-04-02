/*
   @Time : 2021/3/30 8:29 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : tools_url_shortener.go
   @Package : structs
   @Description: 工具 短链接生成器 ( 长链接转短链接 )
*/

package structs

import "github.com/jinzhu/gorm"

// ToolsUrlShortener 短链接生成器 生成记录表
type ToolsUrlShortener struct {
	gorm.Model
	CreatedUserID uint   `gorm:"not null"` // 创建用户 ID
	UpdatedUserID uint   `gorm:"not null"` // 最终修改用户 ID
	CustomID      string // 自定义 ID
	URL           string `gorm:"not null" binding:"required"` // 原始链接
}

// ToolsUrlShortenerRedirectLog 短链接生成器 跳转记录表
type ToolsUrlShortenerRedirectLog struct {
	gorm.Model
	UrlID         uint   // 短链 ID
	CustomID      string // 自定义 ID
	URL           string // 原始链接
	RequestProto  string // 协议版本
	RequestMethod string // 请求方法
	RequestUrl    string // 请求链接
	RequestBody   string // 请求体
	IP            string // 用户 IP
	UserAgent     string // 用户代理
	Status        string // 跳转状态
}
