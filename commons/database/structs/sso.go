/*
   @Time : 2020/11/6 4:26 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : sso
   @Software: GoLand
   @Description: 结构体 单点登陆
*/

package structs

import "github.com/jinzhu/gorm"

// 单点登陆 登陆模块
type SingleSignOnLoginModule struct {
	gorm.Model
	CreatedUserID uint   `gorm:"not null"`                                                   // 创建用户 ID
	UpdatedUserID uint   `gorm:"not null"`                                                   // 最终修改用户 ID
	CRMEID        string `gorm:"not null;column:crm_eid" json:"EID" binding:"required"`      // CRM 活动编码
	CRMEFID       string `gorm:"not null;column:crm_ef_id" json:"EFID" binding:"required"`   // CRM 活动表单 ID
	CRMEFSID      string `gorm:"not null;column:crm_ef_sid" json:"EFSID" binding:"required"` // CRM 活动表单 SID
	Term          uint   `gorm:"not null" json:"Term" binding:"required"`                    // 验证码有效期, 分钟
	Platform      uint   `gorm:"not null" json:"Platform" binding:"required"`                // 发信平台
	Sign          string `gorm:"not null" json:"Sign" binding:"required"`                    // 发信签名 ( 仅腾讯云短信接口可配置, 中公短信平台的签名不可修改, 但是此处的配置与登陆模块 title 联动 )
	TemplateID    uint   `gorm:"not null" json:"TemplateID"`                                 // 发信模板 ID
}

// 单点登陆 验证码
type SingleSignOnVerificationCode struct {
	gorm.Model
	Phone string `gorm:"not null"` // 手机号码
	Term  uint   `gorm:"not null"` // 有效期, 分钟
	Code  uint   `gorm:"not null"` // 验证码
	// 平台日志
	SourceIP string // 客户 IP
}

// 单点登陆 用户
type SingleSignOnUser struct {
	gorm.Model
	Phone string `gorm:"not null"` // 手机号码
}

// 单点登陆 会话
type SingleSignOnSession struct {
	gorm.Model
	MID   uint   `gorm:"not null" json:"MID" binding:"required"`   // 单点登陆模块 ID
	Phone string `gorm:"not null" json:"Phone" binding:"required"` // 客户手机号码
	Code  uint   `json:"Code"`                                     // 验证码
	// 后缀信息
	ActualSuffix  string `json:"Suffix"`   // 调用接口时使用的后缀
	CurrentSuffix string `gorm:"not null"` // 最终使用的后缀
	// CRM 推送配置
	CRMEFSID   string `gorm:"not null;column:crm_ef_sid"` // CRM 活动表单 SID
	CRMChannel uint   `gorm:"not null"`                   // CRM 所属渠道
	CRMOCode   uint   `gorm:"not null;column:crm_o_code"` // CRM 所属组织代码
	CRMUID     uint   `gorm:"not null"`                   // CRM 用户ID
	// CRM 推送配置 可选字段
	CustomerName       string `json:"CustomerName"`       // 客户姓名
	CustomerIdentityID uint   `json:"CustomerIdentityID"` // 客户身份 ID, 来自 CRM 中的客户身份字典
	CustomerColleage   string `json:"CustomerColleage"`   // 客户毕业院校
	CustomerMayor      string `json:"CustomerMayor"`      // 客户专业
	Remark             string `json:"Remark"`             // 备注
	// 平台日志
	SourceIP string // 用户 IP
}

// 单点登陆 后缀
type SingleSignOnSuffix struct {
	gorm.Model
	CreatedUserID uint   `gorm:"not null"`                                                         // 创建用户 ID
	UpdatedUserID uint   `gorm:"not null"`                                                         // 最终修改用户 ID
	Suffix        string `gorm:"not null;primaryKey" json:"Suffix" binding:"required"`             // 后缀 ( 19课堂 个人后缀 ), 为了避免重复, 设置为主键
	Name          string `gorm:"not null" json:"Name" binding:"required"`                          // 后缀名称
	CRMUser       string `gorm:"not null" json:"CRMUser" binding:"required"`                       // CRM 用户名
	CRMUID        uint   `gorm:"not null" json:"CRMUID" binding:"required"`                        // CRM 用户ID
	CRMOID        uint   `gorm:"not null;column:crm_oid" json:"CRMOID" binding:"required"`         // 所属组织 ID
	CRMChannel    uint   `gorm:"not null" json:"CRMChannel" binding:"required"`                    // 所属渠道
	NTalkerGID    string `gorm:"not null;column:ntalker_gid" json:"NTalkerGID" binding:"required"` // 小能咨询组
}

// 单点登陆 CRM 组织
type SingleSignOnOrganization struct {
	gorm.Model
	CreatedUserID uint   `gorm:"not null"`                                // 创建用户 ID
	UpdatedUserID uint   `gorm:"not null"`                                // 最终修改用户 ID
	FID           uint   `gorm:"not null" json:"FID" binding:"required"`  // 父节点 ID
	Code          uint   `gorm:"not null" json:"Code" binding:"required"` // 组织代码
	Name          string `gorm:"not null" json:"Name" binding:"required"` // 组织名称
}

// 单点登陆 CRM 循环分配日志
type SingleSignOnCRMRoundLog struct {
	gorm.Model
	MID    uint   `gorm:"not null"`                // 单点登陆模块 ID
	Phone  string `gorm:"not null"`                // 客户手机号码
	CRMOID uint   `gorm:"not null;column:crm_oid"` // 所属组织 ID
}

// 单点登陆 错误日志表
type SingleSignOnErrorLog struct {
	gorm.Model
	Phone      string `gorm:"not null"`            // 客户手机号码
	MID        uint   `gorm:"not null"`            // 单点登陆模块 ID
	CRMChannel uint   `gorm:"not null"`            // CRM 所属渠道
	CRMUID     uint   `gorm:"not null"`            // CRM 用户ID
	CRMOCode   uint   `gorm:"not null"`            // CRM 所属组织代码
	Error      string `gorm:"not null;size:10000"` // 错误内容, 错误内容可能比较多, 所以字段长度需要配置的比较长
}

// 单点登陆 推送日志表
type SingleSignOnPushLog struct {
	gorm.Model
	Phone string `gorm:"not null" json:"Phone" binding:"required"` // 客户手机号码
	// 后缀信息
	ActualSuffix  string `json:"Suffix"`   // 调用接口时使用的后缀
	CurrentSuffix string `gorm:"not null"` // 最终使用的后缀
	// CRM 推送配置
	CRMEFSID   string `gorm:"not null;column:crm_ef_sid"` // CRM 活动表单 SID
	CRMChannel uint   `gorm:"not null"`                   // CRM 所属渠道
	CRMOCode   uint   `gorm:"not null;column:crm_o_code"` // CRM 所属组织代码
	CRMUID     uint   `gorm:"not null"`                   // CRM 用户ID
	// CRM 推送配置 可选字段
	CustomerName       string `json:"CustomerName"`       // 客户姓名
	CustomerIdentityID uint   `json:"CustomerIdentityID"` // 客户身份 ID, 来自 CRM 中的客户身份字典
	CustomerColleage   string `json:"CustomerColleage"`   // 客户毕业院校
	CustomerMayor      string `json:"CustomerMayor"`      // 客户专业
	Remark             string `json:"Remark"`             // 备注
}
