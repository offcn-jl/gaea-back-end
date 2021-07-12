/*
   @Time : 2020/11/5 11:26 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : unit_test_tool
   @Software: GoLand
   @Description: 单元测试工具
*/

package utt

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"net/http/httptest"
	"os"
	"time"
)

var (
	ORM                      *gorm.DB                   // 测试用 ORM 对象
	HttpTestResponseRecorder *httptest.ResponseRecorder // Http 上下文
	GinTestContext           *gin.Context               // Gin 上下文
	FakePhoneCount           uint                       // 假号码生成计数，可以有效防止生成重复的号码
)

// tableList 单元测试需要使用的数据库表列表
var tableList = []interface{}{
	// System 系统
	structs.SystemConfig{},
	structs.SystemUser{},
	structs.SystemUserLoginFailLog{},
	structs.SystemSession{},
	structs.SystemRole{},
	// SingleSignOn 单点登陆
	structs.SingleSignOnLoginModule{},
	structs.SingleSignOnVerificationCode{},
	structs.SingleSignOnUser{},
	structs.SingleSignOnSession{},
	structs.SingleSignOnSuffix{},
	structs.SingleSignOnOrganization{},
	structs.SingleSignOnCRMRoundLog{},
	structs.SingleSignOnErrorLog{},
	structs.SingleSignOnPushLog{},
	// MiniProgram 小程序
	structs.MiniProgram{},
	structs.MiniProgramAccessToken{},
	structs.MiniProgramPhotoProcessingConfig{},
}

// CreatORM 单元测试工具 创建 ORM
func CreatORM() {
	// 使用正确的 DSN 初始化 MYSQL 客户端
	ORM, _ = gorm.Open("mysql", os.Getenv("UNIT_TEST_MYSQL_DSN_GAEA"))
}

// InitORM 单元测试工具 初始化 ORM
func InitORM() {
	for _, table := range tableList {
		ORM.AutoMigrate(table)
	}
}

// CloseORM 单元测试工具 关闭 ORM
func CloseORM() {
	for _, table := range tableList {
		ORM.DropTableIfExists(table)
	}

	// 关闭 ORM 的连接
	ORM.Close()
}

// InitTest 单元测试工具 初始化测试数据并创建测试上下文
func InitTest() {
	// 初始化数据库
	CreatORM()
	for _, table := range tableList {
		ORM.DropTableIfExists(table)
	}
	InitORM()

	// 初始化 HTTP 测试所需的上下文
	HttpTestResponseRecorder = httptest.NewRecorder()
	GinTestContext, _ = gin.CreateTestContext(HttpTestResponseRecorder)

	// 初始化测试所需的数据

	// CRM 组织信息
	// 省级分校
	ORM.Create(&structs.SingleSignOnOrganization{Model: gorm.Model{ID: 1}, Code: 22, Name: "吉林分校"})
	// 地市分校 1
	ORM.Create(&structs.SingleSignOnOrganization{Model: gorm.Model{ID: 2}, FID: 1, Code: 2290, Name: "吉林长春分校"})
	// 地市分校 2
	ORM.Create(&structs.SingleSignOnOrganization{Model: gorm.Model{ID: 3}, FID: 1, Code: 2305, Name: "吉林市分校"})

	// 后缀信息
	// 默认后缀 ( ID = 1 )
	ORM.Create(&structs.SingleSignOnSuffix{Model: gorm.Model{ID: 1}, Suffix: "default", CRMUser: "default", CRMUID: 32431 /* 齐* */, CRMOID: 1 /* 吉林分校 */, CRMChannel: 7 /* 19 课堂 ( 网推 ) */, NTalkerGID: "NTalkerGID"})
	// 已删除, 但是依旧有效 ( 未到达配置的删除时间 ) 的后缀
	tmpTime := time.Now().Add(8760 * time.Hour) // 一年后
	ORM.Create(&structs.SingleSignOnSuffix{Model: gorm.Model{ID: 2, DeletedAt: &tmpTime}, Suffix: "test", CRMUser: "test", CRMUID: 123 /* 高** */, CRMOID: 2 /* 吉林长春分校 */, CRMChannel: 22 /* 户外推广 ( 市场 ) */, NTalkerGID: "NTalkerGID"})
	// 已删除, 并且已经失效 ( 到达删除时间 ) 的后缀
	tmpTime = time.Now().Add(-8760 * time.Hour) // 一年前
	ORM.Create(&structs.SingleSignOnSuffix{Model: gorm.Model{ID: 3, DeletedAt: &tmpTime}, Suffix: "expired", CRMUser: "expired", CRMUID: 123 /* 高** */, CRMOID: 2 /* 吉林长春分校 */, CRMChannel: 22 /* 户外推广 ( 市场 ) */, NTalkerGID: "NTalkerGID"})

	// 登陆模块信息
	ORM.Create(&structs.SingleSignOnLoginModule{Model: gorm.Model{ID: 10001}, CRMEID: "HD202010142576", CRMEFID: 56975, CRMEFSID: "f905e07b2bff94d564ac1fa41022a633", Sign: "中公教育"})

	// 单点登陆用户
	ORM.Create(&structs.SingleSignOnUser{Phone: "17888666688"})

	// 单点登陆验证码发送记录
	// 验证码正确, 但是已经失效
	ORM.Create(&structs.SingleSignOnVerificationCode{Model: gorm.Model{CreatedAt: time.Now().Add(-1 * time.Hour)}, Phone: "17888886666", Term: 5, Code: 9999})
	// 验证码正确, 并且有效
	ORM.Create(&structs.SingleSignOnVerificationCode{Phone: "17866886688", Term: 5, Code: 9999})

	return
}

// ResetContext 单元测试工具 重置上下文
func ResetContext() {
	HttpTestResponseRecorder = httptest.NewRecorder()
	GinTestContext, _ = gin.CreateTestContext(HttpTestResponseRecorder)
}

// GetFakePhone 单元测试工具 获取测试用的非重复假号码
func GetFakePhone() string {
	FakePhoneCount += 1
	return "1868648" + fmt.Sprintf("%04d", FakePhoneCount)
}
