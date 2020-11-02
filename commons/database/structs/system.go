/*
   @Time : 2020/10/31 3:20 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : system
   @Software: GoLand
*/

package structs

import (
	"github.com/jinzhu/gorm"
)

type SystemConfig struct {
	gorm.Model
	DisableDebug bool // 关闭调试模式, 由于 bool 类型的默认初始值为 false 为了在没有初始化成功的情况下默认开启调试, 所以使用 禁用调试 替代 开启调试
}
