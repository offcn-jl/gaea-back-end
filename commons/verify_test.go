/*
   @Time : 2020/11/6 4:04 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : unit_test_tools
   @Software: GoLand
*/

package commons

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// 测试 Phone 函数是否可以用来验证手机号码是否有效
func TestPhone(t *testing.T) {
	Convey("测试 Phone 函数是否可以用来验证手机号码是否有效", t, func() {
		So(Verify().Phone("17866668888"), ShouldBeTrue)      // 178 号段
		So(Verify().Phone("17166668888"), ShouldBeTrue)      // 171 号段
		So(Verify().Phone("+8617166668888"), ShouldBeTrue)   // 带国家码 +86
		So(Verify().Phone("008617166668888"), ShouldBeTrue)  // 带国家码 0086
		So(Verify().Phone("27866668888"), ShouldBeFalse)     // 287 号段
		So(Verify().Phone("+0117866668888"), ShouldBeFalse)  // 带国家码 +01
		So(Verify().Phone("000117866668888"), ShouldBeFalse) // 带国家码 0001
	})
}
