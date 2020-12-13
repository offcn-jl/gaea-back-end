/*
   @Time : 2020/12/5 2:36 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : auth
   @Software: GoLand
   @Description: 权限相关
*/

package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
)

// GetUserInfo 从上下文中获取用户信息
func GetUserInfo(c *gin.Context) structs.SystemUser {
	// 校验当前用户所在用户组是否具有管理目标用户组的权限 即目标用户组是否是当前用户所在用户组的子用户组
	if userInfo, exists := c.Get("UserInfo"); !exists {
		return structs.SystemUser{}
	} else {
		return userInfo.(structs.SystemUser)
	}
}
