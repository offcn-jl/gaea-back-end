/*
   @Time : 2020/12/5 2:36 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : utils
   @Description: 工具库
*/

package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"strconv"
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

// GetCurrentRoleIDAndSubordinateRolesIDList 获取当前用户所属角色及下级角色的 ID 列表
func GetCurrentRoleIDAndSubordinateRolesIDList(c *gin.Context) []uint {
	// 从 Gin 的上下文中取出角色信息并判断是否存在角色信息
	if roleInfo, exists := c.Get("RoleInfo"); !exists {
		// 不存在, 返回一个空列表
		return nil
	} else {
		// 存在, 查询列表
		list := []uint{roleInfo.(structs.SystemRole).ID}
		list = append(list, getSubordinateRoles(list[0])...)
		return list
	}
}

// getSubordinateRoles 工具函数 递归取出各级下属角色的信息
func getSubordinateRoles(superiorID uint) (list []uint) {
	logger.DebugToJson("上级角色的 ID", superiorID)

	// 取出下属角色列表
	rows, _ := orm.MySQL.Gaea.Raw("SELECT id FROM system_roles WHERE superior_id = ?", superiorID).Rows()
	defer rows.Close()
	for rows.Next() {
		var id uint
		rows.Scan(&id)
		list = append(list, id)
	}

	logger.DebugToJson("下级角色的 ID 列表", list)

	// 递归遍历下属角色列表
	for _, value := range list {
		logger.DebugToString("将要开始获取信息的上级级角色的 ID", value)
		// 递归获取下属角色的下属角色列表
		list = append(list, getSubordinateRoles(value)...)
	}

	return list
}

// StringToInt string 转int
// 摘自: http://www.57mz.com/programs/golang/52.html , 该文中还有 string 转 time 函数
func StringToInt(str string) int {
	i, e := strconv.Atoi(str)
	if e != nil {
		return 0
	}
	return i
}
