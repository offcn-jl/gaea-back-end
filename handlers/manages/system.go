/*
   @Time : 2020/12/3 9:50 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : system
   @Software: GoLand
   @Description: 系统服务
*/

package manages

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/encrypt"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/verify"
	"github.com/satori/go.uuid"
	"net/http"
	"time"
)

// SystemGetRSAPublicKey 获取 RSA 公钥
func SystemGetRSAPublicKey(c *gin.Context) {
	c.String(http.StatusOK, config.Get().RSAPublicKey)
}

// SystemLogin 进行用户登陆操作
func SystemLogin(c *gin.Context) {
	requestJsonMap := struct {
		Username string `json:"Username" binding:"required"` // 用户名
		Password string `json:"Password" binding:"required"` // 密码
		MisToken string `json:"MisToken" binding:"required"` // Mis 口令码
	}{}
	// 绑定参数
	if err := c.ShouldBindJSON(&requestJsonMap); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 校验 Mis 口令码
	if pass, err := verify.MisToken(requestJsonMap.MisToken); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("校验 Mis 口令码失败", err))
		return
	} else if !pass {
		c.JSON(http.StatusForbidden, response.Message("Mis 口令码不正确"))
		return
	}

	// 使用用户名到数据库中取出用户的密码 (经过 RSA 加密)
	userInfo := structs.SystemUser{}
	orm.MySQL.Gaea.Where("user_name = ?", requestJsonMap.Username).Find(&userInfo)
	// 校验用户是否存在
	if userInfo.Username == "" {
		c.JSON(http.StatusForbidden, response.Message("用户不存在或已经被禁用"))
		return
	}

	// 获取 24h 登陆失败次数，登陆失败次数超过 5 次则拒绝登陆
	// 获取 SystemUserLoginFailLog 中 UID = 要登陆的 UID AND CreateAt > 当前时间-1d的数据的数量
	loginFailCount := 0
	orm.MySQL.Gaea.Model(structs.SystemUserLoginFailLog{}).Where("user_id = ? AND created_at > ?", userInfo.ID, time.Now().AddDate(0, 0, -1)).Count(&loginFailCount)
	if loginFailCount > 5 {
		c.JSON(http.StatusForbidden, response.Message("用户 24 小时内连续登陆失败 5 次, 已经被暂时冻结, 24 小时后自动解冻"))
		return
	}

	// 将传送来的密码进行 RSA 解密
	if DecryptedPasswordInRequest, err := encrypt.RSADecrypt(requestJsonMap.Password); err != nil {
		// RSA 解密失败
		c.JSON(http.StatusBadRequest, response.Error("请求中的用户密码 RSA 解密失败", err))
		return
	} else {
		// 将数据库中的密码进行 RSA 解密
		if DecryptedPasswordInDatabase, err := encrypt.RSADecrypt(userInfo.Password); err != nil {
			// RSA 解密失败
			c.JSON(http.StatusInternalServerError, response.Error("数据库中的用户密码 RSA 解密失败", err))
			return
		} else {
			// 进行比较
			if string(DecryptedPasswordInDatabase) != string(DecryptedPasswordInRequest) {
				// 记录登陆失败
				orm.MySQL.Gaea.Create(&structs.SystemUserLoginFailLog{UserID: userInfo.ID, Password: string(DecryptedPasswordInRequest), SourceIp: c.ClientIP()})
				c.JSON(http.StatusForbidden, response.Message("密码不正确, 已经登陆失败 "+fmt.Sprint(loginFailCount+1)+" 次"))
				return
			}

			// 生成会话 UUID
			// 使用随机生成的 UUID 作为会话识别代码, 而不使用自增的 ID 作为会话识别代码, 原因是自增的代码可以很容易的被猜测到, 存在很大的被仿冒的风险
			uuidString := uuid.Must(uuid.NewV4(), nil).String()

			// 将会话保存到数据库中
			orm.MySQL.Gaea.Create(&structs.SystemSession{UUID: uuidString, UserID: userInfo.ID, MisToken: requestJsonMap.MisToken, LastRequestAt: time.Now(), LastSourceIP: c.ClientIP()})

			// 返回 UUID
			c.JSON(http.StatusOK, response.Data(uuidString))
		}
	}
}

// SystemLogout 进行退出 ( 销毁会话 ) 操作
// 本接口不需要验证 Mis Token 是否依旧有效, 所以不使用会话检查中间件对会话进行检查, 但是不使用会话检查中间件并不影响从 Header 中获取 UUID , 将 UUID 参数从 Path 中获取修改为从请求头中获取
func SystemLogout(c *gin.Context) {
	orm.MySQL.Gaea.Where("uuid = ?", c.GetHeader("Authorization")[5:]).Delete(structs.SystemSession{})
	c.JSON(http.StatusOK, response.Success)
}

// SystemUpdateMisToken 进行更新 Mis 口令码操作
// 本接口不需要验证 Mis Token 是否依旧有效, 所以不使用会话检查中间件对会话进行检查, 但是不使用会话检查中间件并不影响从 Header 中获取 UUID , 将 UUID 参数从 Path 中获取修改为从请求头中获取
func SystemUpdateMisToken(c *gin.Context) {
	// 判断会话是否过期
	sessionInfo := structs.SystemSession{}
	orm.MySQL.Gaea.Unscoped().Where("uuid = ?", c.GetHeader("Authorization")[5:]).Last(&sessionInfo)
	if sessionInfo.DeletedAt != nil {
		c.JSON(http.StatusUnauthorized, response.Message("会话无效"))
		return
	}

	// 校验 Mis 口令码是否有效
	if pass, err := verify.MisToken(c.Param("MisToken")); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("校验 Mis 口令码失败", err))
	} else if !pass {
		c.JSON(http.StatusForbidden, response.Message("Mis 口令码不正确"))
	} else {
		// 更新 UUID 对应 Session 的 Mis 口令码记录
		orm.MySQL.Gaea.Model(structs.SystemSession{}).Where("uuid = ?", sessionInfo.UUID).Update("mis_token", c.Param("MisToken"))
		c.JSON(http.StatusOK, response.Success)
	}
}
