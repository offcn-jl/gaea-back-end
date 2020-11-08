/*
   @Time : 2020/11/8 8:35 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : suffix
   @Software: GoLand
   @Description: 个人后缀业务的服务接口
*/

package services

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/responses"
	"github.com/offcn-jl/gaea-back-end/handlers/events"
	"net/http"
	"net/url"
	"time"
)

// SuffixGetActive 获取有效的个人后缀 ( 主要用于生成物料等操作 )
func SuffixGetActive(c *gin.Context) {
	if rows, err := orm.MySQL.Gaea.Raw("SELECT suffixes.id,suffixes.suffix,suffixes.`name`,suffixes.crm_user,suffixes.crm_uid,suffixes.crm_channel,suffixes.ntalker_gid,organizations.id,organizations.f_id,organizations.`code`,organizations.`name` FROM single_sign_on_suffixes AS suffixes,single_sign_on_organizations AS organizations WHERE suffixes.deleted_at IS NULL AND suffixes.crm_oid=organizations.id").Order("suffixes.id ASC").Rows(); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, responses.Error("执行 SQL 查询出错", err))
	} else {
		type Result struct {
			ID         uint   // 后缀 ID
			Suffix     string // 后缀
			Name       string // 后缀名称
			CRMUser    string // CRM 用户名
			CRMUID     uint   // CRM 用户 ID
			CRMChannel uint   // CRM 所属渠道
			NTalkerGID string // 小能咨询组
			CRMOID     uint   // CRM 组织 ID
			CRMOFID    uint   // CRM 上级组织 ID
			CRMOCode   uint   // CRM 组织代码
			CRMOName   string // CRM 组织名称
		}
		results := make([]Result, 0)
		for rows.Next() {
			tempResult := Result{}
			if err := rows.Scan(
				&tempResult.ID,
				&tempResult.Suffix,
				&tempResult.Name,
				&tempResult.CRMUser,
				&tempResult.CRMUID,
				&tempResult.CRMChannel,
				&tempResult.NTalkerGID,
				&tempResult.CRMOID,
				&tempResult.CRMOFID,
				&tempResult.CRMOCode,
				&tempResult.CRMOName,
			); err != nil {
				logger.Error(err)
			}
			results = append(results, tempResult)
		}
		if err := rows.Close(); err != nil {
			logger.Error(err)
		}
		c.JSON(http.StatusOK, responses.Data(results))
	}
}

// SuffixGetDeleting 获取即将过期的后缀 用于后缀花名册
func SuffixGetDeleting(c *gin.Context) {
	if rows, err := orm.MySQL.Gaea.Raw("SELECT suffixes.id,suffixes.deleted_at,suffixes.suffix,suffixes.`name`,suffixes.crm_user,suffixes.crm_uid,suffixes.crm_channel,suffixes.ntalker_gid,organizations.id,organizations.f_id,organizations.`code`,organizations.`name` FROM single_sign_on_suffixes AS suffixes,single_sign_on_organizations AS organizations WHERE suffixes.deleted_at> NOW() AND suffixes.crm_oid=organizations.id").Order("suffixes.id ASC").Rows(); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, responses.Error("执行 SQL 查询出错", err))
	} else {
		type Result struct {
			ID         uint      // 后缀 ID
			DeletedAt  time.Time // 禁用时间
			Suffix     string    // 后缀
			Name       string    // 后缀名称
			CRMUser    string    // CRM 用户名
			CRMUID     uint      // CRM 用户 ID
			CRMChannel uint      // CRM 所属渠道
			NTalkerGID string    // 小能咨询组
			CRMOID     uint      // CRM 组织 ID
			CRMOFID    uint      // CRM 上级组织 ID
			CRMOCode   uint      // CRM 组织代码
			CRMOName   string    // CRM 组织名称
		}
		results := make([]Result, 0)
		for rows.Next() {
			tempResult := Result{}
			if err := rows.Scan(
				&tempResult.ID,
				&tempResult.DeletedAt,
				&tempResult.Suffix,
				&tempResult.Name,
				&tempResult.CRMUser,
				&tempResult.CRMUID,
				&tempResult.CRMChannel,
				&tempResult.NTalkerGID,
				&tempResult.CRMOID,
				&tempResult.CRMOFID,
				&tempResult.CRMOCode,
				&tempResult.CRMOName,
			); err != nil {
				logger.Error(err)
			}
			results = append(results, tempResult)
		}
		if err := rows.Close(); err != nil {
			logger.Error(err)
		}
		c.JSON(http.StatusOK, responses.Data(results))
	}
}

// SuffixGetAvailable 获取所有后缀 用于推送数据
func SuffixGetAvailable(c *gin.Context) {
	if rows, err := orm.MySQL.Gaea.Raw("SELECT suffixes.id,suffixes.suffix,suffixes.`name`,suffixes.crm_user,suffixes.crm_uid,suffixes.crm_channel,suffixes.ntalker_gid,organizations.id,organizations.f_id,organizations.`code`,organizations.`name` FROM single_sign_on_suffixes AS suffixes,single_sign_on_organizations AS organizations WHERE suffixes.crm_oid=organizations.id").Order("suffixes.id ASC").Rows(); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, responses.Error("执行 SQL 查询出错", err))
	} else {
		type Result struct {
			ID         uint   // 后缀 ID
			Suffix     string // 后缀
			Name       string // 后缀名称
			CRMUser    string // CRM 用户名
			CRMUID     uint   // CRM 用户 ID
			CRMChannel uint   // CRM 所属渠道
			NTalkerGID string // 小能咨询组
			CRMOID     uint   // CRM 组织 ID
			CRMOFID    uint   // CRM 上级组织 ID
			CRMOCode   uint   // CRM 组织代码
			CRMOName   string // CRM 组织名称
		}
		results := make([]Result, 0)
		for rows.Next() {
			tempResult := Result{}
			if err := rows.Scan(
				&tempResult.ID,
				&tempResult.Suffix,
				&tempResult.Name,
				&tempResult.CRMUser,
				&tempResult.CRMUID,
				&tempResult.CRMChannel,
				&tempResult.NTalkerGID,
				&tempResult.CRMOID,
				&tempResult.CRMOFID,
				&tempResult.CRMOCode,
				&tempResult.CRMOName,
			); err != nil {
				logger.Error(err)
			}
			results = append(results, tempResult)
		}
		if err := rows.Close(); err != nil {
			logger.Error(err)
		}
		c.JSON(http.StatusOK, responses.Data(results))
	}
}

// SuffixPushCRM 推送带有个人后缀的信息到 CRM
func SuffixPushCRM(c *gin.Context) {
	pushInfo := structs.SingleSignOnPushLog{}
	// 绑定数据
	if err := c.ShouldBindJSON(&pushInfo); err != nil {
		// 绑定数据错误
		logger.Error(err)
		c.JSON(http.StatusBadRequest, responses.Json.Invalid(err))
		return
	}

	// 验证手机号码是否有效
	if !commons.Verify().Phone(pushInfo.Phone) {
		c.JSON(http.StatusBadRequest, responses.Phone.Invalid)
		return
	}

	// 检查是否已经进行过推送
	pushInfo4Check := structs.SingleSignOnPushLog{}
	orm.MySQL.Gaea.Where("crm_ef_sid = ? AND phone = ?", pushInfo.CRMEFSID, pushInfo.Phone).Find(&pushInfo4Check)
	if pushInfo4Check.ID != 0 {
		// 已经进行过推送, 跳过后续步骤
		c.JSON(http.StatusOK, responses.Success)
		return
	}

	// 校验后缀
	if pushInfo.ActualSuffix == "" {
		// 后缀未填写, 使用默认后缀配置
		tempSession := structs.SingleSignOnSession{Phone: pushInfo.Phone}
		events.SSOGetDefaultSuffix(&tempSession)
		pushInfo.CRMChannel = tempSession.CRMChannel
		pushInfo.CurrentSuffix = tempSession.CurrentSuffix
		pushInfo.CRMUID = tempSession.CRMUID
		pushInfo.CRMOCode = tempSession.CRMOCode
	} else {
		suffixInfo := structs.SingleSignOnSuffix{}
		orm.MySQL.Gaea.Unscoped().Where("suffix = ?", pushInfo.ActualSuffix).Find(&suffixInfo)
		if suffixInfo.ID == 0 {
			// 后缀无效, 使用默认后缀配置
			tempSession := structs.SingleSignOnSession{Phone: pushInfo.Phone}
			events.SSOGetDefaultSuffix(&tempSession)
			pushInfo.CRMChannel = tempSession.CRMChannel
			pushInfo.CurrentSuffix = tempSession.CurrentSuffix
			pushInfo.CRMUID = tempSession.CRMUID
			pushInfo.CRMOCode = tempSession.CRMOCode
		} else {
			pushInfo.CRMChannel = suffixInfo.CRMChannel
			pushInfo.CRMUID = suffixInfo.CRMUID
			pushInfo.CurrentSuffix = suffixInfo.Suffix
			if suffixInfo.CRMOID > 1 {
				// 配置了 CRMOID 并且不是省级
				organizationInfo := structs.SingleSignOnOrganization{}
				orm.MySQL.Gaea.Where("id = ?", suffixInfo.CRMOID).Find(&organizationInfo)
				pushInfo.CRMOCode = organizationInfo.Code
			} else {
				// 未配置 CRMOID 或者是省级 ( 等于 1 ), 按手机号码归属地分配 CRM 信息
				tempSession := structs.SingleSignOnSession{Phone: pushInfo.Phone}
				events.SSODistributionByPhoneNumber(&tempSession)
				pushInfo.CRMOCode = tempSession.CRMOCode
			}
		}
	}

	// 推送信息到 CRM
	urlObject, _ := url.Parse("https://dc.offcn.com:8443/a.gif")
	// 构建参数 queryObject
	queryObject := urlObject.Query()
	queryObject.Set("sid", pushInfo.CRMEFSID)
	queryObject.Set("mobile", pushInfo.Phone)
	queryObject.Set("channel", fmt.Sprint(pushInfo.CRMChannel))
	queryObject.Set("orgn", fmt.Sprint(pushInfo.CRMOCode))
	if pushInfo.CRMUID != 0 {
		queryObject.Set("owner", fmt.Sprint(pushInfo.CRMUID))
	}
	if pushInfo.CustomerName != "" {
		queryObject.Set("name", pushInfo.CustomerName)
	}
	if pushInfo.CustomerIdentityID != 0 {
		queryObject.Set("khsf", fmt.Sprint(pushInfo.CustomerIdentityID))
	}
	if pushInfo.CustomerColleage != "" {
		queryObject.Set("colleage", pushInfo.CustomerColleage)
	}
	if pushInfo.CustomerMayor != "" {
		queryObject.Set("mayor", pushInfo.CustomerMayor)
	}
	if pushInfo.Remark != "" {
		queryObject.Set("remark", pushInfo.Remark)
	}
	// 发送 GET 请求
	urlObject.RawQuery = queryObject.Encode()
	if getResponse, err := http.Get(urlObject.String()); err != nil {
		// 发送 GET 请求出错
		logger.Error(err)
		// 推送失败, 保存推送失败记录
		orm.MySQL.Gaea.Create(&structs.SingleSignOnErrorLog{
			Phone:      pushInfo.Phone,
			MID:        0, // 0 代表本推送接口
			CRMChannel: pushInfo.CRMChannel,
			CRMUID:     pushInfo.CRMUID,
			CRMOCode:   pushInfo.CRMOCode,
			Error:      "推送接口 > 请求失败 : " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, responses.Error("向 CRM 发起请求失败", err))
		return
	} else {
		if getResponse.StatusCode != 200 {
			// 推送失败, 保存推送失败记录
			orm.MySQL.Gaea.Create(&structs.SingleSignOnErrorLog{
				Phone:      pushInfo.Phone,
				MID:        0, // 0 代表本推送接口
				CRMChannel: pushInfo.CRMChannel,
				CRMUID:     pushInfo.CRMUID,
				CRMOCode:   pushInfo.CRMOCode,
				Error:      "推送接口 > CRM 返回了错误的状态码 : " + fmt.Sprint(getResponse.StatusCode),
			})
			c.JSON(http.StatusInternalServerError, responses.Message("CRM 返回了错误的状态码 : "+fmt.Sprint(getResponse.StatusCode)))
			return
		}

		// 保存会话信息
		orm.MySQL.Gaea.Create(&pushInfo)

		// 成功
		c.JSON(http.StatusOK, responses.Success)
	}
}
