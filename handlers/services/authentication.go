/*
   @Time : 2020/11/19 3:27 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : authentication
   @Software: GoLand
   @Description: 认证服务
*/

package services

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/request"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GetMiniProgramAccessToken 获取微信小程序 AccessToken
// 无论是生产环境还是测试环境，内部还是外部。获取小程序的 AccessToken 都只使用这一个入口
// 因为如果重新获取小程序 AccessToken , 会导致前一个 AccessToken 失效, 在测试环境获取 AccessToken 时可能导致生产环境的 AccessToken 失效
func GetMiniProgramAccessToken(c *gin.Context) {
	// 判断运行环境，如果是测试环境则直接调用生产环境，并将生产环境的响应原样返回
	if c.Request.RequestURI[1:8] != "release" {
		// 当前运行环境为非生产环境，将请求转发到生产环境
		// 获取当前环境的域名或路径，然后修改 test 为 release 进行调用
		if responseData, err := http.Get(c.Request.Host + "/release" + c.Request.RequestURI[1+strings.IndexAny(c.Request.RequestURI[1:], "/"):]); err != nil {
			// 发送 GET 请求出错
			c.JSON(http.StatusInternalServerError, response.Error("发送请求失败", err))
		} else {
			defer responseData.Body.Close() // 函数退出时关闭 body
			// 读取 body
			if responseBytes, err := ioutil.ReadAll(responseData.Body); err != nil {
				c.JSON(http.StatusInternalServerError, response.Error("读取响应失败", err))
			} else {
				c.String(responseData.StatusCode, string(responseBytes))
			}
		}
		return
	}

	// 绑定数据
	query := struct {
		AccessToken string `form:"access-token" binding:"required"`
		AppID       string `form:"app-id" binding:"required"`
	}{}

	if err := c.ShouldBindQuery(&query); err != nil {
		// 绑定数据错误
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Query.Invalid(err))
		return
	}

	// 校验 AccessToken 是否合法
	if query.AccessToken != config.Get().ServicesAccessToken {
		c.JSON(http.StatusBadRequest, response.Message("AccessToken 不正确"))
		return
	}

	// 从小程序配置表中取出小程序配置
	miniProgram := structs.MiniProgram{}
	orm.MySQL.Gaea.Where("app_id = ?", query.AppID).Find(&miniProgram)

	// 校验是否存在对应小程序
	if miniProgram.ID == 0 {
		c.JSON(http.StatusNotFound, response.Message("AppID "+query.AppID+" 对应的小程序不存在"))
		return
	}

	accessTokenInfo := structs.MiniProgramAccessToken{}
	orm.MySQL.Gaea.Where("app_id = ?", miniProgram.AppID).Last(&accessTokenInfo)
	if accessTokenInfo.AccessToken == "" {
		// 调用获取函数获取一条 AccessToken 并返回
		c.JSON(requestMiniProgramAccessToken(miniProgram.AppID, miniProgram.Secret))
	} else {
		now := time.Now()                                                                             // 当前时间
		hh, _ := time.ParseDuration("-" + strconv.FormatInt(accessTokenInfo.ExpiresIn-200, 10) + "s") // 要提前的时间为过期时间减少二百秒
		deadLine := now.Add(hh)
		// 判断是否超过有效期
		if accessTokenInfo.CreatedAt.After(deadLine) {
			// 没有超过则直接返回
			c.JSON(http.StatusOK, response.Data(accessTokenInfo.AccessToken))
		} else {
			// 调用获取函数获取一条 AccessToken 并返回
			c.JSON(requestMiniProgramAccessToken(miniProgram.AppID, miniProgram.Secret))
		}
	}
}

/**
 * requestMiniProgramAccessToken 获取一条新的 微信小程序 Access Token 保存到数据库后返回
 */
func requestMiniProgramAccessToken(appID, secret string) (int, interface{}) {
	// 发送请求获取 Access Token
	// grant_type 固定填写 client_credential
	// appid 小程序唯一凭证，即 AppID，可在「微信公众平台 - 设置 - 开发设置」页中获得。 ( 需要已经成为开发者，且帐号没有异常状态 )
	// secret 小程序唯一凭证密钥，即 AppSecret，获取方式同 AppID
	if responseJsonMap, err := request.GetSendQueryReceiveJson("https://api.weixin.qq.com/cgi-bin/token", map[string]string{"grant_type": "client_credential", "appid": appID, "secret": secret}); err != nil {
		// 请求失败
		return http.StatusInternalServerError, response.Error("发送请求失败", err)
	} else {
		// 处理请求结果
		logger.DebugToJson("responseJsonMap", responseJsonMap)
		if responseJsonMap["errcode"] != nil {
			// 请求出错
			return http.StatusForbidden, response.Message("[ " + fmt.Sprint(responseJsonMap["errcode"]) + " ] " + fmt.Sprint(responseJsonMap["errmsg"]))
		} else {
			// 保存
			data := structs.MiniProgramAccessToken{AppID: appID, AccessToken: fmt.Sprint(responseJsonMap["access_token"]), ExpiresIn: int64(responseJsonMap["expires_in"].(float64))}
			orm.MySQL.Gaea.Create(&data)
			return http.StatusOK, response.Data(responseJsonMap["access_token"])
		}
	}
}
