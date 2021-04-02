/*
   @Time : 2021/1/21 2:50 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : system_manages_roles_and_users_manages_test
   @Description: [ 单元测试 ] 系统管理 - 角色与用户管理
*/

package manages

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/encrypt"
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

// createRolesAndUsers 创建测试用的角色与用户信息
func createRolesAndUsers() {
	orm.MySQL.Gaea.DropTableIfExists(structs.SystemRole{})
	orm.MySQL.Gaea.AutoMigrate(structs.SystemRole{})
	orm.MySQL.Gaea.DropTableIfExists(structs.SystemUser{})
	orm.MySQL.Gaea.AutoMigrate(structs.SystemUser{})
	orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1001}, Name: "1001"})
	orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1002}, Name: "1002", SuperiorID: 1001})
	orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})
	orm.MySQL.Gaea.Create(&structs.SystemUser{Model: gorm.Model{ID: 101}, Name: "101", Username: "101", RoleID: 1002, CreatedUserID: 102, UpdatedUserID: 102})
	orm.MySQL.Gaea.Create(&structs.SystemUser{Model: gorm.Model{ID: 102}, Name: "102", Username: "102", RoleID: 1003, CreatedUserID: 101, UpdatedUserID: 101})
	orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1004}, Name: "1004", Permissions: "[ \"fake-permission-1\", \"fake-permission-2\" ]", CreatedUserID: 101, UpdatedUserID: 102})
	orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1005}, Name: "1005", Permissions: "[ \"fake-permission-3\", \"fake-permission-4\" ]", CreatedUserID: 101, UpdatedUserID: 102, SuperiorID: 1004})
	orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1006}, Name: "1006", Permissions: "[ \"fake-permission-5\", \"fake-permission-6\" ]", CreatedUserID: 101, UpdatedUserID: 102, SuperiorID: 1005})
}

// TestSystemManagesRolesAndUsersManagesRoleCreate 测试 SystemManagesRolesAndUsersManagesRoleCreate 是否可以检查信息是否合规并创建角色
func TestSystemManagesRolesAndUsersManagesRoleCreate(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesRoleCreate 是否可以检查信息是否合规并创建角色", t, func() {
		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"SuperiorID\": 1004,\"Name\":\"fake-name\",\"Permissions\":\"fake-permissions\"}"))

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"给新角色配置的上级角色不是当前用户所属角色, 并且不是当前用户所属角色的下属角色\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1002}, Name: "1002", SuperiorID: 1001})

		// 创建测试用的角色信息
		createRolesAndUsers()

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"SuperiorID\": 1003,\"Name\":\"fake-name\",\"Permissions\":\"fake-permissions\"}"))

		// 测试 反序列化新角色权限为数组失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid character 'k' in literal false (expecting 'l')\",\"Message\":\"反序列化新角色权限为数组失败\"}")

		// 添加请求内容
		//utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"SuperiorID\": 1030,\"Name\":\"fake-name\",\"Permissions\":\"[\\\"fake-permissions\\\"]\"}"))

		// 测试 给新角色配置的上级角色不存在 不会出现这种情况, 上级角色不存在时会在前面的检查权限步骤中被阻止, 保留这段代码仅用作注释
		//utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		//SystemManagesRolesAndUsersManagesRoleCreate(utt.GinTestContext)
		//So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"给新角色配置的上级角色不存在\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"SuperiorID\": 1003,\"Name\":\"fake-name\",\"Permissions\":\"[\\\"fake-permissions\\\"]\"}"))

		// 测试 反序列化上级角色权限为数组失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"unexpected end of JSON input\",\"Message\":\"反序列化上级角色权限为数组失败\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1004}, Name: "1004", SuperiorID: 0})

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"SuperiorID\": 1004,\"Name\":\"fake-name\",\"Permissions\":\"[\\\"fake-permissions\\\"]\"}"))

		// 测试 给新角色配置的权限不是其上级角色的权限子集
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"给新角色配置的权限不是其上级角色的权限子集\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"SuperiorID\": 1004,\"Name\":\"fake-name\",\"Permissions\":\"[\\\"fake-permission-1\\\"]\"}"))

		// 测试 返回创建成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试 创建角色
		checkInfo := structs.SystemRole{}
		orm.MySQL.Gaea.Where("name = 'fake-name'").Find(&checkInfo)
		So(checkInfo.Name, ShouldEqual, "fake-name")
	})
}

// TestSystemManagesRolesAndUsersManagesRoleGetTree 测试 SystemManagesRolesAndUsersManagesRoleGetTree 是否可以获取当前用户所属角色及下属角色树
func TestSystemManagesRolesAndUsersManagesRoleGetTree(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesRoleGet 是否可以获取当前用户所属角色及下属角色树", t, func() {
		// 创建测试用的角色信息
		createRolesAndUsers()

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1004}, Name: "1004", SuperiorID: 0})

		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleGetTree(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"ID\":1004,\"Name\":\"1004\",\"Permissions\":\"[ \\\"fake-permission-1\\\", \\\"fake-permission-2\\\" ]\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"LastUpdatedUser\":\"102\",\"Children\":[{\"ID\":1005,\"Name\":\"1005\",\"Permissions\":\"[ \\\"fake-permission-3\\\", \\\"fake-permission-4\\\" ]\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"LastUpdatedUser\":\"102\",\"Children\":[{\"ID\":1006,\"Name\":\"1006\",\"Permissions\":\"[ \\\"fake-permission-5\\\", \\\"fake-permission-6\\\" ]\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"LastUpdatedUser\":\"102\",\"Children\":[]}]}]},\"Message\":\"Success\"")
	})
}

// TestSystemManagesRolesAndUsersManagesRoleUpdateInfo 测试 SystemManagesRolesAndUsersManagesRoleUpdateInfo 是否可以检查请求是否合规并修改角色信息
func TestSystemManagesRolesAndUsersManagesRoleUpdateInfo(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesRoleUpdateInfo 是否可以检查请求是否合规并修改角色信息", t, func() {
		// 清空 Request
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/", nil)

		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateInfo(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1004"}}
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Name\":\"new-fake-name\",\"Permissions\":\"new-fake-permissions\"}"))

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateInfo(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"角色不是当前用户所属角色的下属角色\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1002}, Name: "1002", SuperiorID: 1001})

		// 创建测试用的角色信息
		createRolesAndUsers()

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1002"}}
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Name\":\"new-fake-name\",\"Permissions\":\"new-fake-permissions\"}"))

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateInfo(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"角色不是当前用户所属角色的下属角色\"}")

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1003"}}
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Name\":\"new-fake-name\",\"Permissions\":\"new-fake-permissions\"}"))

		// 测试 反序列化新角色权限为数组失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateInfo(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid character 'e' in literal null (expecting 'u')\",\"Message\":\"反序列化角色新权限为数组失败\"}")

		// 添加请求内容
		//utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"SuperiorID\": 1030,\"Name\":\"fake-name\",\"Permissions\":\"[\\\"fake-permissions\\\"]\"}"))

		// 测试 给新角色配置的上级角色不存在 不会出现这种情况, 上级角色不存在时会在前面的检查权限步骤中被阻止, 保留这段代码仅用作注释
		//utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		//SystemManagesRolesAndUsersManagesRoleCreate(utt.GinTestContext)
		//So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"给新角色配置的上级角色不存在\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Name\":\"new-fake-name\",\"Permissions\":\"[\\\"new-fake-permissions\\\"]\"}"))

		// 测试 反序列化上级角色权限为数组失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateInfo(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"unexpected end of JSON input\",\"Message\":\"反序列化上级角色权限为数组失败\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1004}, Name: "1004", SuperiorID: 0})

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1005"}}
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Name\":\"new-fake-name\",\"Permissions\":\"[\\\"new-fake-permissions\\\"]\"}"))

		// 测试 给新角色配置的权限不是其上级角色的权限子集
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateInfo(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"给新角色配置的权限不是其上级角色的权限子集\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Name\":\"new-fake-name\",\"Permissions\":\"[\\\"fake-permission-1\\\"]\"}"))

		//测试 返回成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateInfo(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试 修改角色信息
		checkInfo := structs.SystemRole{}
		orm.MySQL.Gaea.Where("id = 1005").Find(&checkInfo)
		So(checkInfo.Name, ShouldEqual, "new-fake-name")
		So(checkInfo.Permissions, ShouldEqual, "[\"fake-permission-1\"]")
	})
}

// TestSystemManagesRolesAndUsersManagesRoleUpdateSuperior 测试 SystemManagesRolesAndUsersManagesRoleUpdateSuperior 是否可以检查请求是否合规并修改角色的上级角色
func TestSystemManagesRolesAndUsersManagesRoleUpdateSuperior(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesRoleUpdateSuperior 是否可以检查请求是否合规并修改角色的上级角色", t, func() {
		// 清空 请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/", nil)
		utt.GinTestContext.Params = gin.Params{}

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateSuperior(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"角色不是当前用户所属角色的下属角色\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1001}, Name: "1001", SuperiorID: 1001})

		// 创建测试用的角色信息
		createRolesAndUsers()

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1003"}, gin.Param{Key: "SuperiorRoleID", Value: "1004"}}

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateSuperior(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"角色新上级不是当前用户所属角色或所属角色的下属角色\"}")

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1003"}, gin.Param{Key: "SuperiorRoleID", Value: "1001"}}

		//测试 返回成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateSuperior(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"unexpected end of JSON input\",\"Message\":\"反序列化角色权限为数组失败\"}")

		// 添加角色权限
		orm.MySQL.Gaea.Debug().Model(structs.SystemRole{}).Where("id = 1003").Update(&structs.SystemRole{Permissions: "[\"fake-permission-1\"]"})

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1003"}, gin.Param{Key: "SuperiorRoleID", Value: "1001"}}

		//测试 反序列化上级角色权限失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateSuperior(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"unexpected end of JSON input\",\"Message\":\"反序列化上级角色权限为数组失败\"}")

		// 添加上级角色权限
		orm.MySQL.Gaea.Debug().Model(structs.SystemRole{}).Where("id = 1001").Update(&structs.SystemRole{Permissions: "[\"fake-permission-2\"]"})

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1003"}, gin.Param{Key: "SuperiorRoleID", Value: "1001"}}

		//测试 测试角色配置的权限不是其新上级的权限的子集
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateSuperior(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"角色 [ 1003 ] 配置的权限中，存在其新上级 [ 1001 ] 未配置的权限\"}")

		// 增加上级角色的权限
		orm.MySQL.Gaea.Debug().Model(structs.SystemRole{}).Where("id = 1001").Update(&structs.SystemRole{Permissions: "[\"fake-permission-1\",\"fake-permission-2\"]"})

		// 添加请求内容
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1003"}, gin.Param{Key: "SuperiorRoleID", Value: "1001"}}

		//测试 返回成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesRoleUpdateSuperior(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试 修改角色信息
		checkInfo := structs.SystemRole{}
		orm.MySQL.Gaea.Where("id = 1003").Find(&checkInfo)
		So(checkInfo.SuperiorID, ShouldEqual, 1001)
	})
}

// TestSystemManagesRolesAndUsersManagesUserCreate 测试 SystemManagesRolesAndUsersManagesUserCreate 是否可以检查信息是否合规并创建用户
func TestSystemManagesRolesAndUsersManagesUserCreate(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesUserCreate 是否可以检查信息是否合规并创建用户", t, func() {
		// 清空 Request
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/", nil)

		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 创建测试用的角色与用户信息
		createRolesAndUsers()

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1001}, Name: "1001", SuperiorID: 0})

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1004,\"Username\":\"fake-username\",\"Password\":\"fake-password\",\"Name\":\"fake-name\"}"))

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"配置给用户的角色不是您的下属角色\"}")

		// 修改用户为下属角色
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1003,\"Username\":\"fake-username\",\"Password\":\"fake-password\",\"Name\":\"fake-name\"}"))

		// 测试 校验用户密码是否可以进行解密
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"illegal base64 data at input byte 4\",\"Message\":\"请求中的用户密码 RSA 解密失败\"}")

		// 创建使用 RSA 加密的密码

		// 生成私钥
		// # https://www.cnblogs.com/PeterXu1997/p/12218553.html
		// # https://blog.csdn.net/chenxing1230/article/details/83757638
		// 生成 RSA 密钥对
		// GenerateKey 函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
		// Reader 是一个全局、共享的密码用强随机数生成器
		privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
		// 通过 x509 标准将得到的 ras 私钥序列化为 ASN.1 的 DER 编码字符串
		X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
		// 构建一个 pem.Block 结构体对象
		privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
		// 初始化用于接收 pem 的 buffer
		bufferPrivate := new(bytes.Buffer)
		// 使用 pem 格式对 x509 输出的内容进行编码
		pem.Encode(bufferPrivate, &privateBlock)

		// 生成公钥
		// X509 对公钥编码
		X509PublicKey, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		//创建一个pem.Block结构体对象
		publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
		// 初始化用于接收 pem 的 buffer
		bufferPublic := new(bytes.Buffer)
		// pem格式编码
		pem.Encode(bufferPublic, &publicBlock)

		// 添加 RSA 配置
		config.Update(utt.ORM, structs.SystemConfig{RSAPublicKey: bufferPublic.String(), RSAPrivateKey: bufferPrivate.String()})

		// 使用前面的步骤中生成的 RSA 公钥对密码进行加密
		encryptedRequestPassword, _ := encrypt.RSAEncrypt([]byte("fake-password"))
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1003,\"Username\":\"fake-username\",\"Password\":\""+encryptedRequestPassword+"\",\"Name\":\"fake-name\"}"))

		// 测试 用户密码复杂度 ( 不需要 因为复杂计算插件有自己的单元测试 )
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"密码中应当包含数字\",\"Message\":\"密码强度不符合要求\"}")

		// 使用前面的步骤中生成的 RSA 公钥对密码进行加密
		encryptedRequestPassword, _ = encrypt.RSAEncrypt([]byte("fake-Password1"))
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1003,\"Username\":\"fake-username-system_manages_roles_and_users_manages_test\",\"Password\":\""+encryptedRequestPassword+"\",\"Name\":\"fake-name\"}"))

		// 添加一条用户记录
		orm.MySQL.Gaea.Create(&structs.SystemUser{Username: "fake-username-system_manages_roles_and_users_manages_test"})

		// 测试 校验工号 ( Username ) 是否重复
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"工号为 fake-username-system_manages_roles_and_users_manages_test 的用户已经存在\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1003,\"Username\":\"fake-username-system_manages_roles_and_users_manages_test-1\",\"Password\":\""+encryptedRequestPassword+"\",\"Name\":\"fake-name\"}"))

		// 测试 返回创建成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试 是否成功创建用户
		checkInfo := structs.SystemUser{}
		orm.MySQL.Gaea.Where("username = 'fake-username-system_manages_roles_and_users_manages_test-1'").Find(&checkInfo)
		So(checkInfo.Password, ShouldEqual, encryptedRequestPassword)
	})
}

// TestSystemManagesRolesAndUsersManagesUserPaginationGet 测试 SystemManagesRolesAndUsersManagesUserPaginationGet 是否可以检查参数是否合规并分页获取用户列表
func TestSystemManagesRolesAndUsersManagesUserPaginationGet(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesUserPaginationGet 是否可以检查参数是否合规并分页获取用户列表", t, func() {
		// 创建测试用的角色与用户信息
		createRolesAndUsers()

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1002"}, gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserPaginationGet(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"查询的角色不是当前用户所属角色或所属角色的下属角色\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1002}, Name: "1002", SuperiorID: 1001})

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserPaginationGet(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":101,\"Username\":\"101\",\"Name\":\"101\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, ",\"LastUpdatedUser\":\"102\",\"DeletedAt\":null}],\"Message\":\"Success\",\"Total\":1}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "RoleID", Value: "1003"}, gin.Param{Key: "Page", Value: "1"}, gin.Param{Key: "Limit", Value: "10"}}

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserPaginationGet(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":102,\"Username\":\"102\",\"Name\":\"102\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, ",\"LastUpdatedUser\":\"101\",\"DeletedAt\":null}],\"Message\":\"Success\",\"Total\":1}")
	})
}

// TestSystemManagesRolesAndUsersManagesUserUpdate 测试 SystemManagesRolesAndUsersManagesUserUpdate 是否可以检查参数是否合规并修改用户信息
func TestSystemManagesRolesAndUsersManagesUserUpdate(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesUserUpdate 是否可以检查参数是否合规并修改用户信息", t, func() {
		// 清空 Request
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/", nil)

		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1004,\"Username\":\"fake-username\",\"Password\":\"fake-password\",\"Name\":\"fake-name\"}"))
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UserID", Value: "100"}}

		// 测试 检查是否存在目标用户
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户不存在\"}")

		// 创建测试用的角色与用户信息
		createRolesAndUsers()

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1004,\"Username\":\"101\",\"Password\":\"fake-password\",\"Name\":\"fake-name\"}"))
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UserID", Value: "102"}}

		// 测试 校验工号 ( Username ) 是否重复
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"工号为 101 的用户已经存在\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 0})

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1004,\"Username\":\"new-fake-name\",\"Password\":\"fake-password\",\"Name\":\"fake-name\"}"))

		// 测试 检查当前用户是否具有操作目标用户所在角色的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户的角色不是您的下属角色\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1001}, Name: "1001", SuperiorID: 0})

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1004,\"Username\":\"new-fake-name\",\"Password\":\"fake-password\",\"Name\":\"fake-name\"}"))

		// 测试 检查当前用户是否有操作目标用户组的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"配置给用户的角色不是您的下属角色\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1002,\"Username\":\"new-fake-name\",\"Password\":\"fake-password\",\"Name\":\"fake-name\"}"))

		// 测试 校验用户密码是否可以进行解密
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"illegal base64 data at input byte 4\",\"Message\":\"请求中的用户密码 RSA 解密失败\"}")

		// 创建使用 RSA 加密的密码

		// 生成私钥
		// # https://www.cnblogs.com/PeterXu1997/p/12218553.html
		// # https://blog.csdn.net/chenxing1230/article/details/83757638
		// 生成 RSA 密钥对
		// GenerateKey 函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
		// Reader 是一个全局、共享的密码用强随机数生成器
		privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
		// 通过 x509 标准将得到的 ras 私钥序列化为 ASN.1 的 DER 编码字符串
		X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
		// 构建一个 pem.Block 结构体对象
		privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
		// 初始化用于接收 pem 的 buffer
		bufferPrivate := new(bytes.Buffer)
		// 使用 pem 格式对 x509 输出的内容进行编码
		pem.Encode(bufferPrivate, &privateBlock)

		// 生成公钥
		// X509 对公钥编码
		X509PublicKey, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		//创建一个pem.Block结构体对象
		publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
		// 初始化用于接收 pem 的 buffer
		bufferPublic := new(bytes.Buffer)
		// pem格式编码
		pem.Encode(bufferPublic, &publicBlock)

		// 添加 RSA 配置
		config.Update(utt.ORM, structs.SystemConfig{RSAPublicKey: bufferPublic.String(), RSAPrivateKey: bufferPrivate.String()})

		// 使用前面的步骤中生成的 RSA 公钥对密码进行加密
		encryptedRequestPassword, _ := encrypt.RSAEncrypt([]byte("fake-password"))
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1002,\"Username\":\"new-fake-name\",\"Password\":\""+encryptedRequestPassword+"\",\"Name\":\"fake-name\"}"))

		// 测试 用户密码复杂度 ( 不需要 因为复杂计算插件有自己的单元测试 )
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"密码中应当包含数字\",\"Message\":\"密码强度不符合要求\"}")

		// 使用前面的步骤中生成的 RSA 公钥对密码进行加密
		encryptedRequestPassword, _ = encrypt.RSAEncrypt([]byte("fake-Password1"))
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"RoleID\": 1002,\"Username\":\"new-fake-name\",\"Password\":\""+encryptedRequestPassword+"\",\"Name\":\"fake-name\"}"))

		// 测试 返回成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试 修改信息成功
		checkInfo := structs.SystemUser{}
		orm.MySQL.Gaea.Where("id = 102").Find(&checkInfo)
		So(checkInfo.Username, ShouldEqual, "new-fake-name")
		So(checkInfo.Name, ShouldEqual, "fake-name")
		So(checkInfo.Password, ShouldEqual, encryptedRequestPassword)
	})
}

// TestSystemManagesRolesAndUsersManagesUserDisable 测试 SystemManagesRolesAndUsersManagesUserDisable 是否可以检查参数是否合规并禁用用户
func TestSystemManagesRolesAndUsersManagesUserDisable(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesUserDisable 是否可以检查参数是否合规并禁用用户", t, func() {
		// 创建测试用的角色与用户信息
		createRolesAndUsers()

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UserID", Value: "100"}}

		// 测试 检查是否存在目标用户
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserDisable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户不存在\"}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UserID", Value: "102"}}

		// 清空 上下文中的当前角色信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{})

		// 测试 检查当前用户是否具有操作目标用户所在角色的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserDisable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户的角色不是您的下属角色\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1001}, Name: "1001", SuperiorID: 0})

		// 测试 返回成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserDisable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试 检查用户是否已经被禁用
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserDisable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")
	})
}

// TestSystemManagesRolesAndUsersManagesUserEnable 测试 SystemManagesRolesAndUsersManagesUserEnable 是否可以检查参数是否合规并启用用户
func TestSystemManagesRolesAndUsersManagesUserEnable(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesUserEnable 是否可以检查参数是否合规并启用用户", t, func() {
		// 创建测试用的角色与用户信息
		createRolesAndUsers()

		// 设置用户状态为已经禁用
		orm.MySQL.Gaea.Where("id = 102").Delete(&structs.SystemUser{})

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UserID", Value: "100"}}

		// 测试 检查是否存在目标用户
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserEnable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户不存在\"}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UserID", Value: "102"}}

		// 清空 上下文中的当前角色信息
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{})

		// 测试 检查当前用户是否具有操作目标用户所在角色的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserEnable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户的角色不是您的下属角色\"}")

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1001}, Name: "1001", SuperiorID: 0})

		// 测试 返回成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserEnable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试 检查用户是否已经被禁用
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesUserEnable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")
	})
}

// TestSystemManagesRolesAndUsersManagesSearchUser 测试 SystemManagesRolesAndUsersManagesSearchUser 是否可以按照预期搜索用户, 如果存在则返回用户所属角色及所在位置
func TestSystemManagesRolesAndUsersManagesSearchUser(t *testing.T) {
	Convey("测试 SystemManagesRolesAndUsersManagesSearchUser 是否可以按照预期搜索用户, 如果存在则返回用户所属角色及所在位置", t, func() {
		// 测试 按照搜索类型及参数进行搜索用户操作 [ 搜索条件配置有误 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesSearchUser(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"搜索条件配置有误\"}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Type", Value: "id"}, gin.Param{Key: "Criteria", Value: "101"}}

		// 测试 按照搜索类型及参数进行搜索用户操作 [ ID ] 获取用户信息在第几页并返回数据
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesSearchUser(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Rank\":0,\"RoleID\":1002},\"Message\":\"Success\"}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Type", Value: "name"}, gin.Param{Key: "Criteria", Value: "101"}}

		// 测试 按照搜索类型及参数进行搜索用户操作 [ Name, 姓名 ] 获取用户信息在第几页并返回数据
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesSearchUser(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Rank\":0,\"RoleID\":1002},\"Message\":\"Success\"}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Type", Value: "username"}, gin.Param{Key: "Criteria", Value: "101"}}

		// 测试 按照搜索类型及参数进行搜索用户操作 [ Username, 工号 ] 获取用户信息在第几页并返回数据
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesSearchUser(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Rank\":0,\"RoleID\":1002},\"Message\":\"Success\"}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Type", Value: "username"}, gin.Param{Key: "Criteria", Value: "unknown"}}

		// 测试 检查是否存在目标用户
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesRolesAndUsersManagesSearchUser(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户不存在\"}")
	})
}
