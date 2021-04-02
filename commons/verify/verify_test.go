/*
   @Time : 2020/11/6 4:04 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : unit_test_tools
   @Software: GoLand
   @Description: 验证工具 单元测试
*/

package verify

import (
	"github.com/jarcoal/httpmock"
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/response"
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

// TestPhone 测试 Phone 函数是否可以用来验证手机号码是否有效
func TestPhone(t *testing.T) {
	Convey("测试 Phone 函数是否可以用来验证手机号码是否有效", t, func() {
		So(Phone("17866668888"), ShouldBeTrue)      // 178 号段
		So(Phone("17166668888"), ShouldBeTrue)      // 171 号段
		So(Phone("+8617166668888"), ShouldBeTrue)   // 带国家码 +86
		So(Phone("008617166668888"), ShouldBeTrue)  // 带国家码 0086
		So(Phone("27866668888"), ShouldBeFalse)     // 287 号段
		So(Phone("+0117866668888"), ShouldBeFalse)  // 带国家码 +01
		So(Phone("000117866668888"), ShouldBeFalse) // 带国家码 0001
	})
}

// TestMisToken 测试 MisToken 是否可以校验 MIS 口令码 是否合法
func TestMisToken(t *testing.T) {
	Convey("测试 MisToken 是否可以校验 MIS 口令码 是否合法", t, func() {
		// 测试请求失败的情况
		// 此时未配置接口地址, 会直接返回请求出错
		pass, err := MisToken("")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldContainSubstring, "unsupported protocol scheme \"\"")

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 2, "msg": "时间戳缺失或失效"}))

		// 测试 获取 MIS TOKEN 失败 并返回了错误信息
		pass, err = MisToken("")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldContainSubstring, "时间戳缺失或失效")

		// 修改 httpmock 为不返回错误信息
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 2}))

		// 测试 获取 MIS TOKEN 失败 并没有返回错误信息
		pass, err = MisToken("")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldContainSubstring, "请求 MIS 口令码 失败")

		// 修改 httpmock 为返回成功
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 1, "msg": "SIGN验签成功", "data": "fake-token"}))

		// 测试 MIS Token 错误
		pass, err = MisToken("wrong-fake-token")
		So(pass, ShouldBeFalse)
		So(err, ShouldBeNil)

		// 测试 MIS Token 正确
		pass, err = MisToken("fake-token")
		So(pass, ShouldBeTrue)
		So(err, ShouldBeNil)
	})
}

// TestIsSubordinateRole 测试 IsSubordinateRole 检查是否是下属角色
func TestIsSubordinateRole(t *testing.T) {
	Convey("测试 IsSubordinateRole 检查是否是下属角色", t, func() {
		// 测试 检查传入的上级
		So(IsSubordinateRole(0, 0), ShouldBeFalse)

		// 创建 测试角色信息
		orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1001}, Name: "1001"})
		orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1002}, Name: "1002", SuperiorID: 1001})
		orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1003}, Name: "1003", SuperiorID: 1002})

		// 测试 遍历直属下级角色
		So(IsSubordinateRole(1001, 1002), ShouldBeTrue)

		// 测试 遍历更深层的下级角色
		So(IsSubordinateRole(1001, 1003), ShouldBeTrue)

		// 测试 无匹配结果
		So(IsSubordinateRole(1004, 1003), ShouldBeFalse)
	})
}

// TestPasswordComplexity 测试 PasswordComplexity 检查密码复杂度是否符合标准
func TestPasswordComplexity(t *testing.T) {
	Convey("测试 PasswordComplexity 检查密码复杂度是否符合标准", t, func() {
		// 测试 检查密码长度
		pass, err := PasswordComplexity("")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldEqual, "密码长度不足 8 位")

		// 测试 检查密码中是否包含数字
		pass, err = PasswordComplexity("aaaaaaaa")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldEqual, "密码中应当包含数字")

		// 测试 检查密码中是否包含小写字母
		pass, err = PasswordComplexity("12345678")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldEqual, "密码中应当包含小写字母")

		// 测试 检查密码中是否包含大写字母
		pass, err = PasswordComplexity("1234567a")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldEqual, "密码中应当包含大写字母")

		// 测试 检查密码中是否包含特殊符号
		pass, err = PasswordComplexity("123456aA")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldEqual, "密码中应当包含特殊符号，如 ~!@#$%^&*?_- ")

		// 测试 通过检查
		pass, err = PasswordComplexity("12345aA!")
		So(pass, ShouldBeTrue)
		So(err, ShouldBeNil)
	})
}
