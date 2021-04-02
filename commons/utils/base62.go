/*
   @Time : 2021/4/2 10:42 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : base62.go
   @Package : utils
   @Description: 十进制与六十二进制 ( 数字加大小写字母 ) 转换
*/

package utils

import (
	"math"
)

var base = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

// Base62Encode 将十进制数字转换为六十二进制文本
func Base62Encode(num uint) string {
	baseStr := ""
	for {
		if num <= 0 {
			break
		}
		i := num % 62
		baseStr += base[i]
		num = (num - i) / 62
	}
	return baseStr
}

// Base62Decode 将六十二进制文本转换为十进制数字
func Base62Decode(base62 string) uint {
	rs := uint(0)
	length := len(base62)
	f := make(map[string]uint)
	for index, value := range base {
		f[value] = uint(index)
	}
	for i := 0; i < length; i++ {
		rs += f[string(base62[i])] * uint(math.Pow(62, float64(i)))
	}
	return rs
}
