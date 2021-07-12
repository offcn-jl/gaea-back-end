/*
   @Time : 2021/7/8 2:06 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : mini_program_photo_processing.go
   @Package : manages
   @Description: 小程序 证件照
*/

package manages

import (
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

// MiniProgramPhotoProcessingCreate 新建
func MiniProgramPhotoProcessingCreate(c *gin.Context) {
	// 绑定参数
	requestInfo := structs.MiniProgramPhotoProcessingConfig{}
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 校验参数是否合法
	if !regexp.MustCompile(`^\w{32}$`).MatchString(requestInfo.CRMEventFormSID) {
		c.JSON(http.StatusBadRequest, response.Message("CRM 活动表单 SID 格式不正确"))
		return
	}

	// 添加创建用户及修改用户信息
	userInfo := utils.GetUserInfo(c)
	requestInfo.CreatedUserID = userInfo.ID
	requestInfo.UpdatedUserID = userInfo.ID

	// 创建记录
	orm.MySQL.Gaea.Create(&requestInfo)

	// 返回创建成功
	c.JSON(http.StatusOK, response.Success)
}

// MiniProgramPhotoProcessingUpdate 修改
func MiniProgramPhotoProcessingUpdate(c *gin.Context) {
	// 绑定参数
	requestInfo := structs.MiniProgramPhotoProcessingConfig{}
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 校验参数是否合法
	if !regexp.MustCompile(`^\w{32}$`).MatchString(requestInfo.CRMEventFormSID) {
		c.JSON(http.StatusBadRequest, response.Message("CRM 活动表单 SID 格式不正确"))
		return
	}

	// 取出数据库中的记录用于进行检查
	checkInfo := structs.MiniProgramPhotoProcessingConfig{}
	orm.MySQL.Gaea.Unscoped().Where("id = ?", requestInfo.ID).Find(&checkInfo)

	// 取出创建用户及其的所属角色
	createdUserInfo := structs.SystemUser{}
	orm.MySQL.Gaea.Unscoped().Where("id = ?", checkInfo.CreatedUserID).Find(&createdUserInfo)

	// 从上下文中取出角色信息
	roleInfo := utils.GetRoleInfo(c)

	// 检查当前用户所属角色是否有操作目标角色的权限
	if roleInfo.ID != createdUserInfo.RoleID && !verify.IsSubordinateRole(roleInfo.ID, createdUserInfo.RoleID) {
		c.JSON(http.StatusForbidden, response.Message("该证件照配置的创建角色不是当前用户所属角色, 并且不是当前用户所属角色的下属角色"))
		return
	}

	// 更新
	orm.MySQL.Gaea.Unscoped().Model(structs.MiniProgramPhotoProcessingConfig{}).Where("id = ?", requestInfo.ID).Update(map[string]interface{}{"deleted_at": requestInfo.DeletedAt, "name": requestInfo.Name, "project": requestInfo.Project, "crm_event_form_id": requestInfo.CRMEventFormID, "crm_event_form_s_id": requestInfo.CRMEventFormSID, "millimeter_width": requestInfo.MillimeterWidth, "millimeter_height": requestInfo.MillimeterHeight, "pixel_width": requestInfo.PixelWidth, "pixel_height": requestInfo.PixelHeight, "background_colors": requestInfo.BackgroundColors, "description": requestInfo.Description, "hot": requestInfo.Hot})

	// 操作成功
	c.JSON(http.StatusOK, response.Success)
}

// MiniProgramPhotoProcessingGetList 获取列表 带搜索
func MiniProgramPhotoProcessingGetList(c *gin.Context) {
	// 矫正分页参数
	page := utils.StringToInt(c.Query("page"))
	if page == 0 {
		page = 1
	}
	limit := utils.StringToInt(c.Query("limit"))
	if limit == 0 || limit > 100 {
		limit = 10
	}

	// 获取当前用户所属角色及下级角色的 ID 列表
	currentRoleIDAndSubordinateRolesIDList := utils.GetCurrentRoleIDAndSubordinateRolesIDList(c)

	// 定义用于保存信息的切片
	var paginationList []struct {
		ID              uint       // ID
		CreatedAt       time.Time  // 创建时间
		CreatedUser     string     // 创建用户
		CreatedRole     string     // 创建用户所属角色
		LastUpdatedAt   time.Time  // 最终修改时间
		LastUpdatedUser string     // 最终修改用户
		LastUpdatedRole string     // 最终修改用户所属角色
		DeletedAt       *time.Time // 禁用时间
		Operational     bool       // 当前用户是否可以操作

		Name             string // 照片处理名称
		Project          string // 项目
		CRMEventFormID   uint   // CRM 活动表单 ID
		CRMEventFormSID  string // CRM 活动表单 SID
		MillimeterWidth  uint   // MM 毫米 宽度
		MillimeterHeight uint   // MM 毫米 高度
		PixelWidth       uint   // PX 像素 宽度
		PixelHeight      uint   // PX 像素 高度
		BackgroundColors string // 背景色列表
		Description      string // 备注
		Hot              bool   // 是否热门
	}

	// 定义用于保存总量的变量
	total := 0

	// 定义查询基本查询语句
	sql := "SELECT mini_program_photo_processing_configs.id,mini_program_photo_processing_configs.created_at,created_user.`name` AS created_user,created_role.`name` AS created_role,mini_program_photo_processing_configs.updated_at AS last_updated_at,last_updated_user.`name` AS last_updated_user,last_updated_role.`name` AS last_updated_role,mini_program_photo_processing_configs.deleted_at,(SELECT TRUE WHERE created_user.role_id IN (?)) AS operational,mini_program_photo_processing_configs.`name`,mini_program_photo_processing_configs.project,mini_program_photo_processing_configs.crm_event_form_id,mini_program_photo_processing_configs.crm_event_form_s_id,mini_program_photo_processing_configs.millimeter_width,mini_program_photo_processing_configs.millimeter_height,mini_program_photo_processing_configs.pixel_width,mini_program_photo_processing_configs.pixel_height,mini_program_photo_processing_configs.background_colors,mini_program_photo_processing_configs.description,mini_program_photo_processing_configs.hot FROM mini_program_photo_processing_configs,system_users AS created_user,system_roles AS created_role,system_users AS last_updated_user,system_roles AS last_updated_role WHERE mini_program_photo_processing_configs.created_user_id=created_user.id AND created_user.role_id=created_role.id AND mini_program_photo_processing_configs.updated_user_id=last_updated_user.id AND last_updated_user.role_id=last_updated_role.id"

	// 判断是否配置了搜索条件
	if (len(c.Request.URL.Query()) > 0 && c.Query("page") == "" && c.Query("limit") == "") || (len(c.Request.URL.Query()) > 1 && (c.Query("page") == "" || c.Query("limit") == "")) || len(c.Request.URL.Query()) > 2 {
		// 配置了搜索条件
		// 按照传入的查询条件拼接查询语句
		criteria := ""
		// ID [ 强匹配 ]
		if c.Query("id") != "" {
			criteria += " AND mini_program_photo_processing_configs.id = " + c.Query("id")
		}
		// 名称 [ 模糊搜索 ]
		if c.Query("name") != "" {
			criteria += " AND mini_program_photo_processing_configs.`name` LIKE '%" + c.Query("name") + "%'"
		}
		// 项目 [ 强匹配 ]
		if c.Query("project") != "" {
			criteria += " AND mini_program_photo_processing_configs.project = '" + c.Query("project") + "'"
		}
		// CRM 活动表单 ID [ 强匹配 ]
		if c.Query("crm-event-form-id") != "" {
			criteria += " AND mini_program_photo_processing_configs.crm_event_form_id = " + c.Query("crm-event-form-id")
		}
		// CRM 活动表单 SID [ 强匹配 ]
		if c.Query("crm-event-form-sid") != "" {
			// 校验参数是否合法
			if !regexp.MustCompile(`^\w{32}$`).MatchString(c.Query("crm-event-form-sid")) {
				c.JSON(http.StatusBadRequest, response.Message("CRM 活动表单 SID 格式不正确"))
				return
			}
			criteria += " AND mini_program_photo_processing_configs.crm_event_form_s_id = '" + c.Query("crm-event-form-sid") + "'"
		}
		// 是否热门 [ 强匹配 ]
		if c.Query("hot") != "" {
			criteria += " AND mini_program_photo_processing_configs.hot = " + c.Query("hot")
		}
		// 创建用户 条件三合一 ( ID [ 强匹配 ]; 工号 [ 模糊搜索 ]; 姓名 [ 模糊搜索 ] )
		if c.Query("created-user") != "" {
			criteria += " AND ( created_user.id = '" + c.Query("created-user") + "' OR created_user.username LIKE '%" + c.Query("created-user") + "%' OR created_user.name LIKE '%" + c.Query("created-user") + "%' )"
		}
		if criteria == "" {
			c.JSON(http.StatusNotFound, response.Message("搜索条件配置有误"))
			return
		}

		// 按照参数执行分页查询
		orm.MySQL.Gaea.Raw(sql+criteria, currentRoleIDAndSubordinateRolesIDList).Offset((page - 1) * limit).Limit(limit).Order("id DESC").Scan(&paginationList)

		// 按照参数查询总量
		if c.Query("created-user") != "" {
			orm.MySQL.Gaea.Debug().Raw("SELECT COUNT(*) FROM mini_program_photo_processing_configs, system_users AS created_user WHERE mini_program_photo_processing_configs.created_user_id = created_user.id" + criteria).Count(&total)
		} else {
			// 搜索条件不包含于用户相关的内容, 直接执行查询
			// 使用 criteria[5:] 的原因是为了裁剪查询语句开头的 ' AND '
			orm.MySQL.Gaea.Debug().Unscoped().Model(structs.MiniProgramPhotoProcessingConfig{}).Where(criteria[5:]).Count(&total)
		}
	} else {
		// 没有配置搜索条件
		// 按照参数执行分页查询
		orm.MySQL.Gaea.Raw(sql, currentRoleIDAndSubordinateRolesIDList).Offset((page - 1) * limit).Limit(limit).Order("id DESC").Scan(&paginationList)

		// 按照参数查询总量
		orm.MySQL.Gaea.Unscoped().Model(structs.MiniProgramPhotoProcessingConfig{}).Count(&total)
	}

	// 返回分页数据
	c.JSON(http.StatusOK, response.PaginationData(paginationList, total))
}
