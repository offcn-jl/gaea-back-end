/*
   @Time : 2021/4/2 10:57 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : base62_test.go
   @Package : utils
   @Description: 单元测试 十进制与六十二进制 ( 数字加大小写字母 ) 转换
*/

package utils

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestBase62Encode 测试 Base62Encode 是否可以将十进制数字转换为六十二进制文本
func TestBase62Encode(t *testing.T) {
	Convey("测试 Base62Encode 是否可以将十进制数字转换为六十二进制文本", t, func() {
		So(Base62Encode(10), ShouldEqual, "k")
		So(Base62Encode(1000), ShouldEqual, "iq")
		So(Base62Encode(100000), ShouldEqual, "5aA")
		So(Base62Encode(10000000), ShouldEqual, "uC8P")
		So(Base62Encode(1000000000), ShouldEqual, "qQ4Pfb")
	})
}

// TestBase62Decode 测试 Base62Decode 将六十二进制文本转换为十进制数字
func TestBase62Decode(t *testing.T) {
	Convey("测试 Base62Decode 将六十二进制文本转换为十进制数字", t, func() {
		So(Base62Decode("k"), ShouldEqual, 10)
		So(Base62Decode("iq"), ShouldEqual, 1000)
		So(Base62Decode("5aA"), ShouldEqual, 100000)
		So(Base62Decode("uC8P"), ShouldEqual, 10000000)
		So(Base62Decode("qQ4Pfb"), ShouldEqual, 1000000000)
	})
}
