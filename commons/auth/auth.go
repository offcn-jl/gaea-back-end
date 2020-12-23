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

// GetUserInfo 从 Gin 的上下文中获取用户信息
func GetUserInfo(c *gin.Context) structs.SystemUser {
	// 从 Gin 的上下文中取出用户信息并判断是否存在角色信息
	if userInfo, exists := c.Get("UserInfo"); !exists {
		// 不存在, 返回一个空结构体, 业务侧需要根据结构体是否为空来判断是否成功获取到信息
		return structs.SystemUser{}
	} else {
		// 存在, 返回信息
		return userInfo.(structs.SystemUser)
	}
}

// GetRoleInfo 从 Gin 的上下文中获取角色信息
func GetRoleInfo(c *gin.Context) structs.SystemRole {
	// 从 Gin 的上下文中取出角色信息并判断是否存在角色信息
	if roleInfo, exists := c.Get("RoleInfo"); !exists {
		// 不存在, 返回一个空结构体, 业务侧需要根据结构体是否为空来判断是否成功获取到信息
		return structs.SystemRole{}
	} else {
		// 存在, 返回信息
		return roleInfo.(structs.SystemRole)
	}
}
