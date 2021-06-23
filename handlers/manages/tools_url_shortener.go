/*
   @Time : 2021/3/30 8:54 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : tools_url_shortener.go
   @Package : manages
   @Description: 工具 短链接生成器 ( 长链接转短链接 )
*/

package manages

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/utils"
	"github.com/offcn-jl/gaea-back-end/commons/verify"
	"net/http"
	"regexp"
	"time"
)

// ToolsUrlShortenerCreateShortLink 新建短链接
func ToolsUrlShortenerCreateShortLink(c *gin.Context) {
	// 绑定参数
	requestInfo := structs.ToolsUrlShortener{}
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 检查是否已经存在短链记录
	checkInfo := structs.ToolsUrlShortener{}
	orm.MySQL.Gaea.Where("url = ?", requestInfo.URL).Find(&checkInfo)
	if checkInfo.ID != 0 {
		c.JSON(http.StatusOK, response.Data(map[string]interface{}{
			"Repetitive":       true,               // 是否重复
			"ShortUrlID":       checkInfo.ID,       // 短链接 ID , 作为短链接使用需要将其转换为 62 进制
			"ShortUrlCustomID": checkInfo.CustomID, // 短链接自定义 ID
		}))
		return
	}

	// 检查是否配置了自定义 ID
	if requestInfo.CustomID != "" {
		// 检查自定义 ID 是否已经被使用
		orm.MySQL.Gaea.Where("custom_id = ?", requestInfo.CustomID).Find(&checkInfo)
		if checkInfo.ID != 0 {
			c.JSON(http.StatusForbidden, response.Message("自定义短链接 "+requestInfo.CustomID+" 已经被使用, 记录 ID 为 "+fmt.Sprint(checkInfo.ID)))
			return
		}
		// https://www.zhihu.com/question/24474922
		// 校验是否带有一个 HTTP 安全符号 ( 避免与 Base62 的 ID 重复 ), 安全符号包括 -_.!~*'()  共 9 个
		// 使用正则匹配判断自定义 ID 的内容用是否符合要求
		if !regexp.MustCompile(`^[0-9A-Za-z-_.!~*'()]+$`).MatchString(requestInfo.CustomID) {
			c.JSON(http.StatusBadRequest, gin.H{"Code": -1, "Error": "自定义短链接格式不正确, 仅可输入数字、大写字母、小写字母、部分英文符号【 -_.!~*'() 】"})
			return
		}
		// 使用正则匹配判断自定义 ID 中是否带有至少一个 HTTP 安全符号
		if !regexp.MustCompile(`[-_.!~*'()]+`).MatchString(requestInfo.CustomID) {
			c.JSON(http.StatusBadRequest, gin.H{"Code": -1, "Error": "自定义短链接中应当至少包含符号 -_.!~*'() 中的一个"})
			return
		}
	}

	// 添加创建用户及修改用户信息
	userInfo := utils.GetUserInfo(c)
	requestInfo.CreatedUserID = userInfo.ID
	requestInfo.UpdatedUserID = userInfo.ID

	// 创建记录
	orm.MySQL.Gaea.Create(&requestInfo)

	// 返回创建成功
	c.JSON(http.StatusOK, response.Data(map[string]interface{}{
		"Repetitive":       false,                // 是否重复
		"ShortUrlID":       requestInfo.ID,       // 短链接 ID , 作为短链接使用需要将其转换为 62 进制
		"ShortUrlCustomID": requestInfo.CustomID, // 短链接自定义 ID
	}))
}

// ToolsUrlShortenerUpdateShortLink 修改短链接
func ToolsUrlShortenerUpdateShortLink(c *gin.Context) {
	// 绑定参数
	requestInfo := structs.ToolsUrlShortener{}
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 取出数据库中的记录用于进行检查
	checkInfo := structs.ToolsUrlShortener{}
	orm.MySQL.Gaea.Unscoped().Where("id = ?", requestInfo.ID).Find(&checkInfo)

	// 取出短链接创建用户及其的所属角色
	createdUserInfo := structs.SystemUser{}
	orm.MySQL.Gaea.Unscoped().Where("id = ?", checkInfo.CreatedUserID).Find(&createdUserInfo)

	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户是否有操作目标角色的权限
	if roleInfo.ID != createdUserInfo.RoleID && !verify.IsSubordinateRole(roleInfo.ID, createdUserInfo.RoleID) {
		c.JSON(http.StatusForbidden, response.Message("短链接的创建角色不是当前用户所属角色, 并且不是当前用户所属角色的下属角色"))
		return
	}

	// 取出数据库中的记录用于进行检查
	checkInfo1 := structs.ToolsUrlShortener{}
	// 检查是否已经存在短链记录
	orm.MySQL.Gaea.Where("url = ?", requestInfo.URL).Find(&checkInfo1)
	if checkInfo1.ID != 0 && checkInfo1.ID != checkInfo.ID {
		c.JSON(http.StatusForbidden, response.Message("链接 "+requestInfo.URL+" 已经存在转换记录, 记录 ID 为 "+fmt.Sprint(checkInfo1.ID)))
		return
	}

	// 检查是否配置了自定义 ID
	if requestInfo.CustomID != "" {
		checkInfo2 := structs.ToolsUrlShortener{}
		// 检查自定义 ID 是否已经被使用
		orm.MySQL.Gaea.Where("custom_id = ? AND id != ?", requestInfo.CustomID, requestInfo.ID).Find(&checkInfo2)
		if checkInfo2.ID != 0 {
			c.JSON(http.StatusForbidden, response.Message("自定义短链接 "+requestInfo.CustomID+" 已经被使用, 记录 ID 为 "+fmt.Sprint(checkInfo2.ID)))
			return
		}
		// https://www.zhihu.com/question/24474922
		// 校验是否带有一个 HTTP 安全符号 ( 避免与 Base62 的 ID 重复 ), 安全符号包括 -_.!~*'()  共 9 个
		// 使用正则匹配判断自定义 ID 的内容用是否符合要求
		if !regexp.MustCompile(`^[0-9A-Za-z-_.!~*'()]+$`).MatchString(requestInfo.CustomID) {
			c.JSON(http.StatusBadRequest, gin.H{"Code": -1, "Error": "自定义短链接格式不正确, 仅可输入数字、大写字母、小写字母、部分英文符号【 -_.!~*'() 】"})
			return
		}
		// 使用正则匹配判断自定义 ID 中是否带有至少一个 HTTP 安全符号
		if !regexp.MustCompile(`[-_.!~*'()]+`).MatchString(requestInfo.CustomID) {
			c.JSON(http.StatusBadRequest, gin.H{"Code": -1, "Error": "自定义短链接中应当至少包含符号 -_.!~*'() 中的一个"})
			return
		}
	}

	// 更新
	orm.MySQL.Gaea.Unscoped().Model(structs.ToolsUrlShortener{}).Where("id = ?", requestInfo.ID).Update(map[string]interface{}{"deleted_at": requestInfo.DeletedAt, "custom_id": requestInfo.CustomID, "url": requestInfo.URL})

	// 返回成功
	c.JSON(http.StatusOK, response.Success)
}

// 定义信息结构
type shortLinkInfo struct {
	ID              uint       // 短链接 ID
	CustomID        string     // 自定义短链接 ID
	URL             string     // 原始链接
	CreatedAt       time.Time  // 创建时间
	CreatedUser     string     // 创建用户
	CreatedRole     string     // 创建用户所属角色
	LastUpdatedAt   time.Time  // 最终修改时间
	LastUpdatedUser string     // 最终修改用户
	LastUpdatedRole string     // 最终修改用户所属角色
	DeletedAt       *time.Time // 禁用时间
	Operational     bool       // 当前用户是否可以操作
}

// ToolsUrlShortenerGetList 获取短链接列表 ( 带搜索 )
func ToolsUrlShortenerGetList(c *gin.Context) {
	// 获取当前用户所属角色及下级角色的 ID 列表
	currentRoleIDAndSubordinateRolesIDList := utils.GetCurrentRoleIDAndSubordinateRolesIDList(c)

	// 定义信息列表
	var shortLinkPaginationList []shortLinkInfo

	// 定义用于保存总量的变量
	total := 0

	// 判断是否配置了搜索条件
	if len(c.Request.URL.Query()) > 0 {
		// 配置了搜索条件
		// 按照传入的查询条件拼接查询语句
		criteria := ""
		// ID [ 强匹配 ]
		if c.Query("id") != "" {
			criteria += " AND tools_url_shorteners.id = " + c.Query("id")
		}
		// 自定义 ID [ 模糊搜索 ]
		if c.Query("custom-id") != "" {
			criteria += " AND tools_url_shorteners.custom_id LIKE '%" + c.Query("custom-id") + "%'"
		}
		// 链接 [ 模糊搜索 ]
		if c.Query("url") != "" {
			criteria += " AND tools_url_shorteners.url LIKE '%" + c.Query("url") + "%'"
		}
		// 创建用户 条件三合一 ( ID [ 强匹配 ]; 工号 [ 模糊搜索 ]; 姓名 [ 模糊搜索 ] )
		if c.Query("created-user") != "" {
			criteria += " AND ( created_user.id = " + c.Query("created-user") + " OR created_user.username LIKE '%" + c.Query("created-user") + "%' OR created_user.name LIKE '%" + c.Query("created-user") + "%' )"
		}
		if criteria == "" {
			c.JSON(http.StatusNotFound, response.Message("搜索条件配置有误"))
			return
		}

		// 按照参数执行分页查询
		orm.MySQL.Gaea.Raw("SELECT tools_url_shorteners.id,tools_url_shorteners.custom_id,tools_url_shorteners.url,tools_url_shorteners.created_at,created_user.`name` AS created_user,created_role.`name` AS created_role,tools_url_shorteners.updated_at,last_updated_user.`name` AS last_updated_user,last_updated_role.`name` AS last_updated_role,tools_url_shorteners.deleted_at,(SELECT TRUE WHERE created_user.role_id IN (?)) AS operational FROM tools_url_shorteners,system_users AS created_user,system_roles AS created_role,system_users AS last_updated_user,system_roles AS last_updated_role WHERE tools_url_shorteners.created_user_id=created_user.id AND created_user.role_id=created_role.id AND tools_url_shorteners.updated_user_id=last_updated_user.id AND last_updated_user.role_id=last_updated_role.id"+criteria, currentRoleIDAndSubordinateRolesIDList).Offset((utils.StringToInt(c.Param("Page")) - 1) * utils.StringToInt(c.Param("Limit"))).Limit(utils.StringToInt(c.Param("Limit"))).Order("id DESC").Scan(&shortLinkPaginationList)

		// 按照参数查询总量
		if c.Query("created-user") != "" {
			orm.MySQL.Gaea.Debug().Raw("SELECT COUNT(*) FROM tools_url_shorteners, system_users AS created_user WHERE tools_url_shorteners.created_user_id = created_user.id" + criteria).Count(&total)
		} else {
			// 搜索条件不包含于用户相关的内容, 直接执行查询
			// 使用 criteria[5:] 的原因是为了裁剪查询语句开头的 ' AND '
			orm.MySQL.Gaea.Debug().Unscoped().Model(structs.ToolsUrlShortener{}).Where(criteria[5:]).Count(&total)
		}
	} else {
		// 没有配置搜索条件
		// 按照参数执行分页查询
		orm.MySQL.Gaea.Raw("SELECT tools_url_shorteners.id,tools_url_shorteners.custom_id,tools_url_shorteners.url,tools_url_shorteners.created_at,created_user.`name` AS created_user,created_role.`name` AS created_role,tools_url_shorteners.updated_at,last_updated_user.`name` AS last_updated_user,last_updated_role.`name` AS last_updated_role,tools_url_shorteners.deleted_at,(SELECT TRUE WHERE created_user.role_id IN (?)) AS operational FROM tools_url_shorteners,system_users AS created_user,system_roles AS created_role,system_users AS last_updated_user,system_roles AS last_updated_role WHERE tools_url_shorteners.created_user_id=created_user.id AND created_user.role_id=created_role.id AND tools_url_shorteners.updated_user_id=last_updated_user.id AND last_updated_user.role_id=last_updated_role.id", currentRoleIDAndSubordinateRolesIDList).Offset((utils.StringToInt(c.Param("Page")) - 1) * utils.StringToInt(c.Param("Limit"))).Limit(utils.StringToInt(c.Param("Limit"))).Order("id DESC").Scan(&shortLinkPaginationList)

		// 按照参数查询总量
		orm.MySQL.Gaea.Unscoped().Model(structs.ToolsUrlShortener{}).Count(&total)
	}

	// 返回分页数据
	c.JSON(http.StatusOK, response.PaginationData(shortLinkPaginationList, total))
}
