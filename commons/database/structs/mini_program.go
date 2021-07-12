/*
   @Time : 2020/11/19 3:56 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : mini_program
   @Software: GoLand
   @Description: 结构体 小程序
*/

package structs

import "github.com/jinzhu/gorm"

// MiniProgram 小程序配置
type MiniProgram struct {
	gorm.Model
	AppID  string `gorm:"not null"` // 小程序 AppID
	Secret string `gorm:"not null"` // 小程序密钥
}

// MiniProgramAccessToken 小程序 Access Token
type MiniProgramAccessToken struct {
	gorm.Model
	AppID       string `gorm:"not null"` // 小程序 AppID
	AccessToken string `gorm:"not null"` // 获取到的访问令牌
	ExpiresIn   int64  `gorm:"not null"` // 访问令牌有效时间，单位：秒。目前是7200秒之内的值
}

// MiniProgramPhotoProcessingConfig 小程序 照片处理 配置
type MiniProgramPhotoProcessingConfig struct {
	gorm.Model
	CreatedUserID    uint   `gorm:"not null"`                    // 创建用户 ID
	UpdatedUserID    uint   `gorm:"not null"`                    // 最终修改用户 ID
	Name             string `gorm:"not null" binding:"required"` // 照片处理名称
	Project          string `gorm:"not null" binding:"required"` // 项目
	CRMEventFormID   uint   `gorm:"not null" binding:"required"` // CRM 活动表单 ID
	CRMEventFormSID  string `gorm:"not null" binding:"required"` // CRM 活动表单 SID
	MillimeterWidth  uint   `gorm:"not null" binding:"required"` // MM 毫米 宽度
	MillimeterHeight uint   `gorm:"not null" binding:"required"` // MM 毫米 高度
	PixelWidth       uint   `gorm:"not null" binding:"required"` // PX 像素 宽度
	PixelHeight      uint   `gorm:"not null" binding:"required"` // PX 像素 高度
	BackgroundColors string `gorm:"not null" binding:"required"` // 背景色列表
	Description      string // 备注
	Hot              bool   // 是否热门
}
