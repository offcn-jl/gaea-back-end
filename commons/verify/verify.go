/*
   @Time : 2020/11/6 4:05 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : verify
   @Software: GoLand
   @Description: 验证工具
*/

package verify

import "regexp"

// Phone 用来验证手机号码是否有效
// 中国(严谨), 根据工信部2019年最新公布的手机号段
// 摘自 https://any86.github.io/any-rule/
// 去除首尾的 /
func Phone(phone string) bool {
	regular := `^(?:(?:\+|00)86)?1(?:(?:3[\d])|(?:4[5-7|9])|(?:5[0-3|5-9])|(?:6[5-7])|(?:7[0-8])|(?:8[\d])|(?:9[1|8|9]))\d{8}$`
	return regexp.MustCompile(regular).MatchString(phone)
}
