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
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// 测试 Phone 函数是否可以用来验证手机号码是否有效
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
