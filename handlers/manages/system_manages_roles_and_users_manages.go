/*
   @Time : 2021/1/21 11:30 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : system_manages_roles_and_users_manages
   @Description: 系统管理 - 角色与用户管理
*/

package manages

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/encrypt"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/utils"
	"github.com/offcn-jl/gaea-back-end/commons/verify"
	"net/http"
	"time"
)

// SystemManagesRolesAndUsersManagesRoleCreate 创建角色
func SystemManagesRolesAndUsersManagesRoleCreate(c *gin.Context) {
	// 绑定参数
	requestInfo := structs.SystemRole{}
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否有操作目标用户组的权限
	if roleInfo.ID != requestInfo.SuperiorID && !verify.IsSubordinateRole(roleInfo.ID, requestInfo.SuperiorID) {
		c.JSON(http.StatusForbidden, response.Message("给新角色配置的上级角色不是当前用户所属角色, 并且不是当前用户所属角色的下属角色"))
		return
	}

	// 校验新角色的权限是否是上级角色权限的子集
	// 定义新角色权限数组
	var newRolePermissions []string
	// 反序列化新角色权限为数组
	if err := json.Unmarshal([]byte(requestInfo.Permissions), &newRolePermissions); err != nil {
		// 返回错误提示
		c.JSON(http.StatusBadRequest, response.Error("反序列化新角色权限为数组失败", err))
		return
	} else {
		// 取出上级角色信息
		superior := structs.SystemRole{}
		orm.MySQL.Gaea.Where("id = ?", requestInfo.SuperiorID).Find(&superior)
		//if superior.ID == 0 { // 不会出现这种情况, 上级角色不存在时会在前面的检查权限步骤中被阻止, 保留这段代码仅用作注释
		//	// 上级角色不存在，返回错误信息
		//	c.JSON(http.StatusBadRequest, response.Message("给新角色配置的上级角色不存在"))
		//	return
		//}
		// 定义上级角色权限数组
		var superiorPermissions []string
		// 反序列化上级角色权限为数组
		if err := json.Unmarshal([]byte(superior.Permissions), &superiorPermissions); err != nil {
			// 返回错误提示
			c.JSON(http.StatusInternalServerError, response.Error("反序列化上级角色权限为数组失败", err))
			return
		} else {
			// 遍历给新角色配置的所有权限, 检查是否是上级角色的权限的子集
			for newRolePermissionIndex := range newRolePermissions {
				has := false
				for superiorPermissionIndex := range superiorPermissions {
					if newRolePermissions[newRolePermissionIndex] == superiorPermissions[superiorPermissionIndex] {
						has = true
						break
					}
				}
				if !has {
					// 返回错误提示
					c.JSON(http.StatusForbidden, response.Message("给新角色配置的权限不是其上级角色的权限子集"))
					return
				}
			}
		}
	}

	// 添加创建用户及修改用户信息
	userInfo := utils.GetUserInfo(c)
	requestInfo.CreatedUserID = userInfo.ID
	requestInfo.UpdatedUserID = userInfo.ID

	// 创建角色
	orm.MySQL.Gaea.Create(&requestInfo)

	// 返回创建成功
	c.JSON(http.StatusOK, response.Success)
}

// RoleNode 角色信息节点
type RoleNode struct {
	ID              uint
	Name            string
	Permissions     string
	CreatedAt       time.Time
	CreatedUser     string
	LastUpdatedAt   time.Time
	LastUpdatedUser string
	Children        []RoleNode
}

// getSubordinateRoles 取出下属角色的信息
func getSubordinateRoles(roleNode *RoleNode) {
	// 取出下属角色列表
	orm.MySQL.Gaea.Raw("SELECT role.id,role.`name`,role.permissions,role.created_at,created_user.`name` AS created_user,role.updated_at AS last_updated_at,last_updated_user.`name` AS last_updated_user FROM system_roles AS role,system_users AS created_user,system_users AS last_updated_user WHERE role.superior_id=? AND role.created_user_id=created_user.id AND role.updated_user_id=last_updated_user.id", roleNode.ID).Scan(&roleNode.Children)

	// 递归遍历下属角色列表
	for index := range roleNode.Children {
		// 递归获取下属角色的下属角色列表
		getSubordinateRoles(&roleNode.Children[index])
	}
}

// SystemManagesRolesAndUsersManagesRoleGetTree 获取当前用户所属角色及下属角色树
func SystemManagesRolesAndUsersManagesRoleGetTree(c *gin.Context) {
	// 取出用户本人所属角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 初始化角色树
	roleTree := RoleNode{}

	// 取出当前用户所属角色的信息
	orm.MySQL.Gaea.Raw("SELECT role.id,role.`name`,role.permissions,role.created_at,created_user.`name` AS created_user,role.updated_at AS last_updated_at,last_updated_user.`name` AS last_updated_user FROM system_roles AS role,system_users AS created_user,system_users AS last_updated_user WHERE role.id=? AND role.created_user_id=created_user.id AND role.updated_user_id=last_updated_user.id", roleInfo.ID).Scan(&roleTree)

	// 取出下属角色的信息
	getSubordinateRoles(&roleTree)

	// 返回数据
	c.JSON(http.StatusOK, response.Data(roleTree))
}

// SystemManagesRolesAndUsersManagesRoleUpdateInfo 修改角色信息
func SystemManagesRolesAndUsersManagesRoleUpdateInfo(c *gin.Context) {
	// 定义请求信息的类型
	type requestType struct {
		Name        string `json:"Name" binding:"required"`
		Permissions string `json:"Permissions" binding:"required"`
	}

	// 定义请求信息
	requestInfo := requestType{}

	// 绑定数据
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否有操作目标用户组的权限
	if !verify.IsSubordinateRole(roleInfo.ID, uint(utils.StringToInt(c.Param("RoleID")))) {
		c.JSON(http.StatusForbidden, response.Message("角色不是当前用户所属角色的下属角色"))
		return
	}

	// 取出用于检查的信息 ( 不需要这个步骤, 因为前面的 检查当前用户是否有操作目标用户组的权限 逻辑中, 隐式的完成了检查角色是否存在的需求, 保留这段代码仅用作备注 )
	//checkRoleInfo := structs.SystemRole{}
	//orm.MySQL.Gaea.Where("id = ?", requestInfo.ID).Find(&checkRoleInfo)
	// 检查需要修改的角色是否存在
	//if checkRoleInfo.ID == 0 {
	//	c.JSON(http.StatusNotFound, response.Message("角色不存在"))
	//	return
	//}

	// 需要校验角色的新权限是否是上级角色权限的子集
	// 定义新角色权限数组
	var newRolePermissions []string
	// 反序列化新角色权限为数组
	if err := json.Unmarshal([]byte(requestInfo.Permissions), &newRolePermissions); err != nil {
		// 返回错误提示
		c.JSON(http.StatusBadRequest, response.Error("反序列化角色新权限为数组失败", err))
		return
	} else {
		// 取出上级角色信息
		role := structs.SystemRole{}
		orm.MySQL.Gaea.Where("id = ?", c.Param("RoleID")).Find(&role)
		superior := structs.SystemRole{}
		orm.MySQL.Gaea.Where("id = ?", role.SuperiorID).Find(&superior)
		//if superior.ID == 0 { // 不会出现这种情况, 上级角色不存在时会在前面的检查权限步骤中被阻止, 保留这段代码仅用作注释
		//	// 上级角色不存在，返回错误信息
		//	c.JSON(http.StatusBadRequest, response.Message("给新角色配置的上级角色不存在"))
		//	return
		//}
		// 定义上级角色权限数组
		var superiorPermissions []string
		// 反序列化上级角色权限为数组
		if err := json.Unmarshal([]byte(superior.Permissions), &superiorPermissions); err != nil {
			// 返回错误提示
			c.JSON(http.StatusInternalServerError, response.Error("反序列化上级角色权限为数组失败", err))
			return
		} else {
			// 遍历给新角色配置的所有权限, 检查是否是上级角色的权限的子集
			for newRolePermissionIndex := range newRolePermissions {
				has := false
				for superiorPermissionIndex := range superiorPermissions {
					if newRolePermissions[newRolePermissionIndex] == superiorPermissions[superiorPermissionIndex] {
						has = true
						break
					}
				}
				if !has {
					// 返回错误提示
					c.JSON(http.StatusForbidden, response.Message("给新角色配置的权限不是其上级角色的权限子集"))
					return
				}
			}
		}
	}

	// 获取当前用户信息
	userInfo := utils.GetUserInfo(c)

	// 修改角色信息
	orm.MySQL.Gaea.Model(structs.SystemRole{}).Where("id = ?", c.Param("RoleID")).Update(structs.SystemRole{UpdatedUserID: userInfo.ID, Name: requestInfo.Name, Permissions: requestInfo.Permissions})

	// 返回成功
	c.JSON(http.StatusOK, response.Success)
}

// SystemManagesRolesAndUsersManagesRoleUpdateSuperior 修改角色的上级角色
func SystemManagesRolesAndUsersManagesRoleUpdateSuperior(c *gin.Context) {
	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否有操作目标用户组的权限
	if !verify.IsSubordinateRole(roleInfo.ID, uint(utils.StringToInt(c.Param("RoleID")))) {
		c.JSON(http.StatusForbidden, response.Message("角色不是当前用户所属角色的下属角色"))
		return
	}
	if roleInfo.ID != uint(utils.StringToInt(c.Param("SuperiorRoleID"))) && !verify.IsSubordinateRole(roleInfo.ID, uint(utils.StringToInt(c.Param("SuperiorRoleID")))) {
		c.JSON(http.StatusForbidden, response.Message("角色新上级不是当前用户所属角色或所属角色的下属角色"))
		return
	}

	// 取出用于检查的信息 ( 不需要这个步骤, 因为前面的 检查当前用户是否有操作目标用户组的权限 逻辑中, 隐式的完成了检查角色是否存在的需求, 保留这段代码仅用作备注 )
	//checkRoleInfo := structs.SystemRole{}
	//orm.MySQL.Gaea.Where("id = ?", requestInfo.ID).Find(&checkRoleInfo)
	// 检查需要修改的角色是否存在
	//if checkRoleInfo.ID == 0 {
	//	c.JSON(http.StatusNotFound, response.Message("角色不存在"))
	//	return
	//}

	// 取出上级角色信息
	role := structs.SystemRole{}
	orm.MySQL.Gaea.Where("id = ?", c.Param("RoleID")).Find(&role)
	superior := structs.SystemRole{}
	orm.MySQL.Gaea.Where("id = ?", c.Param("SuperiorRoleID")).Find(&superior)

	// 需要角色的权限是否是上级角色权限的子集
	// 定义角色权限数组
	var rolePermissions []string
	// 反序列化角色权限为数组
	if err := json.Unmarshal([]byte(role.Permissions), &rolePermissions); err != nil {
		// 返回错误提示
		c.JSON(http.StatusBadRequest, response.Error("反序列化角色权限为数组失败", err))
		return
	} else {
		// 定义上级角色权限数组
		var superiorPermissions []string
		// 反序列化上级角色权限为数组
		if err := json.Unmarshal([]byte(superior.Permissions), &superiorPermissions); err != nil {
			// 返回错误提示
			c.JSON(http.StatusInternalServerError, response.Error("反序列化上级角色权限为数组失败", err))
			return
		} else {
			// 遍历给新角色配置的所有权限, 检查是否是上级角色的权限的子集
			for rolePermissionIndex := range rolePermissions {
				has := false
				for superiorPermissionIndex := range superiorPermissions {
					if rolePermissions[rolePermissionIndex] == superiorPermissions[superiorPermissionIndex] {
						has = true
						break
					}
				}
				if !has {
					// 返回错误提示
					c.JSON(http.StatusForbidden, response.Message("角色 [ "+role.Name+" ] 配置的权限中，存在其新上级 [ "+superior.Name+" ] 未配置的权限"))
					return
				}
			}
		}
	}

	// 获取当前用户信息
	userInfo := utils.GetUserInfo(c)

	// 修改角色信息
	orm.MySQL.Gaea.Model(structs.SystemRole{}).Where("id = ?", c.Param("RoleID")).Update(structs.SystemRole{UpdatedUserID: userInfo.ID, SuperiorID: uint(utils.StringToInt(c.Param("SuperiorRoleID")))})

	// 返回成功
	c.JSON(http.StatusOK, response.Success)
}

// SystemManagesRolesAndUsersManagesUserCreate 创建用户
func SystemManagesRolesAndUsersManagesUserCreate(c *gin.Context) {
	// 绑定参数
	requestInfo := structs.SystemUser{}
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否有操作目标用户组的权限
	if !verify.IsSubordinateRole(roleInfo.ID, requestInfo.RoleID) {
		c.JSON(http.StatusForbidden, response.Message("配置给用户的角色不是您的下属角色"))
		return
	}

	// 校验用户密码是否可以进行解密
	if DecryptedPasswordInRequest, err := encrypt.RSADecrypt(requestInfo.Password); err != nil {
		// RSA 解密失败
		c.JSON(http.StatusBadRequest, response.Error("请求中的用户密码 RSA 解密失败", err))
		return
	} else if pass, err := verify.PasswordComplexity(string(DecryptedPasswordInRequest)); !pass {
		// 校验用户密码复杂度
		c.JSON(http.StatusForbidden, response.Error("密码强度不符合要求", err))
		return
	}

	// 校验工号 ( Username ) 是否重复
	checkInfo := structs.SystemUser{}
	orm.MySQL.Gaea.Where("username = ?", requestInfo.Username).Find(&checkInfo)
	if checkInfo.ID != 0 {
		c.JSON(http.StatusForbidden, response.Message("工号为 "+requestInfo.Username+" 的用户已经存在"))
		return
	}

	// 添加创建用户及修改用户信息
	userInfo := utils.GetUserInfo(c)
	requestInfo.CreatedUserID = userInfo.ID
	requestInfo.UpdatedUserID = userInfo.ID

	// 创建用户
	orm.MySQL.Gaea.Create(&requestInfo)

	// 返回成功
	c.JSON(http.StatusOK, response.Success)
}

// SystemManagesRolesAndUsersManagesUserPaginationGet 分页获取用户列表
func SystemManagesRolesAndUsersManagesUserPaginationGet(c *gin.Context) {
	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否有操作目标用户组的权限
	if roleInfo.ID != uint(utils.StringToInt(c.Param("RoleID"))) && !verify.IsSubordinateRole(roleInfo.ID, uint(utils.StringToInt(c.Param("RoleID")))) {
		c.JSON(http.StatusForbidden, response.Message("查询的角色不是当前用户所属角色或所属角色的下属角色"))
		return
	}

	// 定义用户信息结构
	type userInfo struct {
		ID              uint
		Username        string
		Name            string
		CreatedAt       time.Time
		CreatedUser     string
		LastUpdatedAt   time.Time
		LastUpdatedUser string
		DeletedAt       *time.Time
	}
	// 定义用户信息列表
	var userPaginationList []userInfo

	// 按照参数执行分页查询
	orm.MySQL.Gaea.Raw("SELECT users.id,users.username,users.`name`,users.created_at,created_user.`name` AS created_user,users.updated_at AS last_updated_at,last_updated_user.`name` AS last_updated_user,users.deleted_at FROM system_users AS users,system_users AS created_user,system_users AS last_updated_user WHERE users.role_id=? AND users.created_user_id=created_user.id AND users.updated_user_id=last_updated_user.id", c.Param("RoleID")).Offset((utils.StringToInt(c.Param("Page")) - 1) * utils.StringToInt(c.Param("Limit"))).Limit(utils.StringToInt(c.Param("Limit"))).Order("id DESC").Scan(&userPaginationList)

	// 定义用于保存用户总量的变量
	total := 0

	// 按照参数查询用户总量
	orm.MySQL.Gaea.Unscoped().Model(structs.SystemUser{}).Where("role_id = ?", c.Param("RoleID")).Count(&total)

	// 返回分页数据
	c.JSON(http.StatusOK, response.PaginationData(userPaginationList, total))
}

// SystemManagesRolesAndUsersManagesUserUpdate 修改用户信息
func SystemManagesRolesAndUsersManagesUserUpdate(c *gin.Context) {
	// 绑定参数
	requestInfo := structs.SystemUser{}
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 取出数据库中保存的目标用户信息
	checkInfo := structs.SystemUser{}
	orm.MySQL.Gaea.Unscoped().Where("id = ?", c.Param("UserID")).Find(&checkInfo)

	// 检查是否存在目标用户
	if checkInfo.ID == 0 {
		c.JSON(http.StatusNotFound, response.Message("用户不存在"))
		return
	}

	// 校验工号 ( Username ) 是否重复
	checkUsernameInfo := structs.SystemUser{}
	orm.MySQL.Gaea.Unscoped().Where("username = ?", requestInfo.Username).Find(&checkUsernameInfo)
	if checkUsernameInfo.ID != 0 && checkUsernameInfo.ID != uint(utils.StringToInt(c.Param("UserID"))) {
		c.JSON(http.StatusForbidden, response.Message("工号为 "+requestInfo.Username+" 的用户已经存在"))
		return
	}

	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否具有操作目标用户所在角色的权限
	if !verify.IsSubordinateRole(roleInfo.ID, checkInfo.RoleID) {
		c.JSON(http.StatusForbidden, response.Message("用户的角色不是您的下属角色"))
		return
	}

	// 检查当前用户是否有操作目标用户组的权限
	if !verify.IsSubordinateRole(roleInfo.ID, requestInfo.RoleID) {
		c.JSON(http.StatusForbidden, response.Message("配置给用户的角色不是您的下属角色"))
		return
	}

	// 检查是否进行修改密码操作
	if requestInfo.Password != "" {
		// 校验用户密码是否可以进行解密
		if DecryptedPasswordInRequest, err := encrypt.RSADecrypt(requestInfo.Password); err != nil {
			// RSA 解密失败
			c.JSON(http.StatusBadRequest, response.Error("请求中的用户密码 RSA 解密失败", err))
			return
		} else if pass, err := verify.PasswordComplexity(string(DecryptedPasswordInRequest)); !pass {
			// 校验用户密码复杂度
			c.JSON(http.StatusForbidden, response.Error("密码强度不符合要求", err))
			return
		}
	}

	// 添加最终修改用户信息
	userInfo := utils.GetUserInfo(c)
	requestInfo.UpdatedUserID = userInfo.ID

	// 修改用户信息
	orm.MySQL.Gaea.Unscoped().Model(&requestInfo).Where("id = ?", c.Param("UserID")).Update(&requestInfo)

	// 返回成功
	c.JSON(http.StatusOK, response.Success)
}

// SystemManagesRolesAndUsersManagesUserDisable 禁用用户
func SystemManagesRolesAndUsersManagesUserDisable(c *gin.Context) {
	// 取出数据库中保存的目标用户信息
	checkInfo := structs.SystemUser{}
	orm.MySQL.Gaea.Unscoped().Where("id = ?", c.Param("UserID")).Find(&checkInfo)

	// 检查是否存在目标用户
	if checkInfo.ID == 0 {
		c.JSON(http.StatusNotFound, response.Message("用户不存在"))
		return
	}

	// 检查用户是否已经被禁用
	if checkInfo.DeletedAt != nil {
		// 用户已经被禁用, 直接返回成功
		c.JSON(http.StatusOK, response.Success)
		return
	}

	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否具有操作目标用户所在角色的权限
	if !verify.IsSubordinateRole(roleInfo.ID, checkInfo.RoleID) {
		c.JSON(http.StatusForbidden, response.Message("用户的角色不是您的下属角色"))
		return
	}

	// 取出用户信息
	userInfo := utils.GetUserInfo(c)

	// 禁用用户
	orm.MySQL.Gaea.Model(&checkInfo).Updates(map[string]interface{}{"updated_user_id": userInfo.ID, "deleted_at": time.Now()})

	// 返回成功
	c.JSON(http.StatusOK, response.Success)
}

// SystemManagesRolesAndUsersManagesUserEnable 启用用户
func SystemManagesRolesAndUsersManagesUserEnable(c *gin.Context) {
	// 取出数据库中保存的目标用户信息
	checkInfo := structs.SystemUser{}
	orm.MySQL.Gaea.Unscoped().Where("id = ?", c.Param("UserID")).Find(&checkInfo)

	// 检查是否存在目标用户
	if checkInfo.ID == 0 {
		c.JSON(http.StatusNotFound, response.Message("用户不存在"))
		return
	}

	// 检查用户是否已经被启用
	if checkInfo.DeletedAt == nil {
		// 用户已经被启用, 直接返回成功
		c.JSON(http.StatusOK, response.Success)
		return
	}

	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否具有操作目标用户所在角色的权限
	if !verify.IsSubordinateRole(roleInfo.ID, checkInfo.RoleID) {
		c.JSON(http.StatusForbidden, response.Message("用户的角色不是您的下属角色"))
		return
	}

	// 取出用户信息
	userInfo := utils.GetUserInfo(c)

	// 启用用户
	orm.MySQL.Gaea.Unscoped().Model(&checkInfo).Updates(map[string]interface{}{"updated_user_id": userInfo.ID, "deleted_at": nil})

	// 返回成功
	c.JSON(http.StatusOK, response.Success)
}

// SystemManagesRolesAndUsersManagesSearchUser 搜索用户, 如果存在则返回用户所属角色及所在位置
func SystemManagesRolesAndUsersManagesSearchUser(c *gin.Context) {
	userInfo := structs.SystemUser{}
	switch c.Param("Type") {
	case "id":
		orm.MySQL.Gaea.Unscoped().Where("id = ?", c.Param("Criteria")).Find(&userInfo)
	case "name":
		orm.MySQL.Gaea.Unscoped().Where("name = ?", c.Param("Criteria")).Find(&userInfo)
	case "username":
		orm.MySQL.Gaea.Unscoped().Where("username = ?", c.Param("Criteria")).Find(&userInfo)
	default:
		c.JSON(http.StatusNotFound, response.Message("搜索条件配置有误"))
		return
	}

	// 检查是否存在目标用户
	if userInfo.ID == 0 {
		c.JSON(http.StatusNotFound, response.Message("用户不存在"))
		return
	}

	// 获取用户信息在第几页
	rank := 0
	orm.MySQL.Gaea.Unscoped().Model(structs.SystemUser{}).Where("role_id = ? AND id > ?", userInfo.RoleID, userInfo.ID).Count(&rank)

	// 返回数据
	c.JSON(http.StatusOK, response.Data(response.Struct{"RoleID": userInfo.RoleID, "Rank": rank}))
}
