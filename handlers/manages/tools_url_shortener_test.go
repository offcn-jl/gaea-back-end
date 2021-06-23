/*
   @Time : 2021/3/30 3:15 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : tools_url_shortener_test.go
   @Package : manages
   @Description: [ 单元测试 ] 工具 短链接生成器 ( 长链接转短链接 )
*/

package manages

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

// 初始化测试数据并获取测试所需的上下文
func init() {
	utt.InitTest()
	orm.MySQL.Gaea = utt.ORM // 覆盖 orm 库中的 ORM 对象
}

// createTestData 创建测试数据
func createTestData() {
	orm.MySQL.Gaea.DropTableIfExists(structs.ToolsUrlShortener{})
	orm.MySQL.Gaea.AutoMigrate(structs.ToolsUrlShortener{})
	orm.MySQL.Gaea.Create(&structs.ToolsUrlShortener{Model: gorm.Model{ID: 10}, CreatedUserID: 101, UpdatedUserID: 101, CustomID: "custom-id-1", URL: "https://host-1.domain"})
	orm.MySQL.Gaea.Create(&structs.ToolsUrlShortener{Model: gorm.Model{ID: 11}, CreatedUserID: 102, UpdatedUserID: 102, CustomID: "custom-id-2", URL: "https://host-2.domain"})
}

// TestToolsUrlShortenerCreateShortLink 测试 ToolsUrlShortenerCreateShortLink 是否可以新建短链接
func TestToolsUrlShortenerCreateShortLink(t *testing.T) {
	Convey("测试 ToolsUrlShortenerCreateShortLink 是否可以新建短链接", t, func() {
		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"EOF\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 创建测试数据
		createTestData()

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"CustomID\": \"custom-id-1\",\"URL\":\"https://host-1.domain\"}"))

		// 测试 检查是否已经存在短链记录
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Repetitive\":true,\"ShortUrlCustomID\":\"custom-id-1\",\"ShortUrlID\":10},\"Message\":\"Success\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"CustomID\": \"custom-id-1\",\"URL\":\"https://host-3.domain\"}"))

		// 测试 检查自定义 ID 是否已经被使用
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"自定义短链接 custom-id-1 已经被使用, 记录 ID 为 10\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"CustomID\": \"custom-id-1^\",\"URL\":\"https://host-3.domain\"}"))

		// 测试 使用正则匹配判断自定义 ID 的内容用是否符合要求
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Code\":-1,\"Error\":\"自定义短链接格式不正确, 仅可输入数字、大写字母、小写字母、部分英文符号【 -_.!~*'() 】\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"CustomID\": \"customid1\",\"URL\":\"https://host-3.domain\"}"))

		// 测试 使用正则匹配判断自定义 ID 中是否带有至少一个 HTTP 安全符号
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Code\":-1,\"Error\":\"自定义短链接中应当至少包含符号 -_.!~*'() 中的一个\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"CustomID\": \"custom-id-3\",\"URL\":\"https://host-3.domain\"}"))

		// 测试 创建记录并返回创建成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Repetitive\":false,\"ShortUrlCustomID\":\"custom-id-3\",\"ShortUrlID\":12},\"Message\":\"Success\"}")
		// 测试是否创建记录
		checkInfo := structs.ToolsUrlShortener{}
		orm.MySQL.Gaea.Where("custom_id = 'custom-id-3'").Find(&checkInfo)
		So(checkInfo.ID, ShouldEqual, 12)
		So(checkInfo.CustomID, ShouldEqual, "custom-id-3")
		So(checkInfo.URL, ShouldEqual, "https://host-3.domain")
	})
}

// TestToolsUrlShortenerUpdateShortLink 测试 ToolsUrlShortenerUpdateShortLink 是否可以修改短链接
func TestToolsUrlShortenerUpdateShortLink(t *testing.T) {
	Convey("测试 ToolsUrlShortenerUpdateShortLink 是否可以修改短链接", t, func() {
		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerUpdateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"EOF\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 创建测试数据
		createTestData()

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1004}, Name: "1004", SuperiorID: 1004})

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": 11,\"CustomID\": \"custom-id-1\",\"URL\":\"https://host-1.domain\"}"))

		// 测试 检查当前用户是否有操作目标角色的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerUpdateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"短链接的创建角色不是当前用户所属角色, 并且不是当前用户所属角色的下属角色\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1002}, Name: "1002", SuperiorID: 1001})

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": 11,\"CustomID\": \"custom-id-1\",\"URL\":\"https://host-1.domain\"}"))

		// 测试 检查是否已经存在短链记录
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerUpdateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"链接 https://host-1.domain 已经存在转换记录, 记录 ID 为 10\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": 11,\"CustomID\": \"custom-id-1\",\"URL\":\"https://host-3.domain\"}"))

		// 测试 检查自定义 ID 是否已经被使用
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerUpdateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"自定义短链接 custom-id-1 已经被使用, 记录 ID 为 10\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": 11,\"CustomID\": \"custom-id-1^\",\"URL\":\"https://host-3.domain\"}"))

		// 测试 使用正则匹配判断自定义 ID 的内容用是否符合要求
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerUpdateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Code\":-1,\"Error\":\"自定义短链接格式不正确, 仅可输入数字、大写字母、小写字母、部分英文符号【 -_.!~*'() 】\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": 11,\"CustomID\": \"customid1\",\"URL\":\"https://host-3.domain\"}"))

		// 测试 使用正则匹配判断自定义 ID 中是否带有至少一个 HTTP 安全符号
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerUpdateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Code\":-1,\"Error\":\"自定义短链接中应当至少包含符号 -_.!~*'() 中的一个\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": 11,\"CustomID\": \"custom-id-3\",\"URL\":\"https://host-3.domain\"}"))

		// 测试 更新并返回成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerUpdateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")
		// 测试 检查数据库中的记录是否修改成功
		checkInfo := structs.ToolsUrlShortener{}
		orm.MySQL.Gaea.Where("custom_id = 'custom-id-3'").Find(&checkInfo)
		So(checkInfo.ID, ShouldEqual, 11)
		So(checkInfo.CustomID, ShouldEqual, "custom-id-3")
		So(checkInfo.URL, ShouldEqual, "https://host-3.domain")
	})
}

// TestToolsUrlShortenerGetList 测试 ToolsUrlShortenerGetList 是否可以按照预期获取短链接列表 ( 带搜索 )
func TestToolsUrlShortenerGetList(t *testing.T) {
	Convey("测试 ToolsUrlShortenerGetList 是否可以按照预期获取短链接列表 ( 带搜索 )", t, func() {
		// 创建测试数据
		createTestData()

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}

		// 配置上下文中的信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 测试 获取全部短链接列表
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":11,\"CustomID\":\"custom-id-2\",\"URL\":\"https://host-2.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, ",\"LastUpdatedUser\":\"102\",\"LastUpdatedRole\":\"1003\",\"DeletedAt\":null,\"Operational\":true},{\"ID\":10,\"CustomID\":\"custom-id-1\",\"URL\":\"https://host-1.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"101\",\"LastUpdatedRole\":\"1002\",\"DeletedAt\":null,\"Operational\":false}],\"Message\":\"Success\",\"Total\":2}")

		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?access-token=fake-access-token-url-shortener", nil)

		// 测试 搜索条件配置有误
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"搜索条件配置有误\"}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?id=10", nil)
		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}
		// 配置上下文中的信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 测试 按照搜索类型及参数进行搜索操作 ID [ 强匹配 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":10,\"CustomID\":\"custom-id-1\",\"URL\":\"https://host-1.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"101\",\"LastUpdatedRole\":\"1002\",\"DeletedAt\":null,\"Operational\":false}],\"Message\":\"Success\",\"Total\":1}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?custom-id=custom", nil)
		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}
		// 配置上下文中的信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 测试 按照搜索类型及参数进行搜索操作 自定义 ID [ 模糊搜索 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":11,\"CustomID\":\"custom-id-2\",\"URL\":\"https://host-2.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"102\",\"LastUpdatedRole\":\"1003\",\"DeletedAt\":null,\"Operational\":true},{\"ID\":10,\"CustomID\":\"custom-id-1\",\"URL\":\"https://host-1.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"101\",\"LastUpdatedRole\":\"1002\",\"DeletedAt\":null,\"Operational\":false}],\"Message\":\"Success\",\"Total\":2}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?url=.domain", nil)
		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}
		// 配置上下文中的信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 测试 按照搜索类型及参数进行搜索操作 链接 [ 模糊搜索 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":11,\"CustomID\":\"custom-id-2\",\"URL\":\"https://host-2.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"102\",\"LastUpdatedRole\":\"1003\",\"DeletedAt\":null,\"Operational\":true},{\"ID\":10,\"CustomID\":\"custom-id-1\",\"URL\":\"https://host-1.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"101\",\"LastUpdatedRole\":\"1002\",\"DeletedAt\":null,\"Operational\":false}],\"Message\":\"Success\",\"Total\":2}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?created-user=101", nil)
		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}
		// 配置上下文中的信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 测试 按照搜索类型及参数进行搜索操作 创建用户 ID [ 强匹配 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":10,\"CustomID\":\"custom-id-1\",\"URL\":\"https://host-1.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"101\",\"LastUpdatedRole\":\"1002\",\"DeletedAt\":null,\"Operational\":false}],\"Message\":\"Success\",\"Total\":1}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?created-user=101", nil)
		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}
		// 配置上下文中的信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 测试 按照搜索类型及参数进行搜索操作 创建用户 工号 [ 模糊搜索 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":10,\"CustomID\":\"custom-id-1\",\"URL\":\"https://host-1.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"101\",\"LastUpdatedRole\":\"1002\",\"DeletedAt\":null,\"Operational\":false}],\"Message\":\"Success\",\"Total\":1}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?created-user=101", nil)
		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}
		// 配置上下文中的信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 测试 按照搜索类型及参数进行搜索操作 创建用户 姓名 [ 模糊搜索 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":10,\"CustomID\":\"custom-id-1\",\"URL\":\"https://host-1.domain\",\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"101\",\"LastUpdatedRole\":\"1002\",\"DeletedAt\":null,\"Operational\":false}],\"Message\":\"Success\",\"Total\":1}")
	})
}
