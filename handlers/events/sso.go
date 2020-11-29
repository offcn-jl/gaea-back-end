/*
   @Time : 2020/11/6 3:55 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : sso
   @Software: GoLand
   @Description: 单点登陆模块的接口及其辅助函数
*/

package events

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/verify"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20190711"
	"github.com/xluohome/phonedata"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// SSOSendVerificationCode 单点登模块发送验证码接口的处理函数
func SSOSendVerificationCode(c *gin.Context) {
	// 验证手机号码是否有效
	if !verify.Phone(c.Param("Phone")) {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Phone.Invalid)
		return
	}

	// 根据登陆模块 ID, 获取登陆模块的配置
	// 需要使用登陆模块配置中的下发平台、签名、模板 ID
	SSOLoginModuleInfo := structs.SingleSignOnLoginModule{}
	orm.MySQL.Gaea.Where("id = ?", c.Param("MID")).Find(&SSOLoginModuleInfo)
	if SSOLoginModuleInfo.ID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Message("登陆模块配置有误"))
	} else {
		switch SSOLoginModuleInfo.Platform {
		case 1:
			// 使用中公短信下发短信
			sendVerificationCodeByOFFCN(c, SSOLoginModuleInfo.TemplateID, SSOLoginModuleInfo.Term)
		case 2:
			// 使用腾讯云下发短信
			sendVerificationCodeByTencentCloudSMSV2(c, SSOLoginModuleInfo.Sign, SSOLoginModuleInfo.TemplateID, SSOLoginModuleInfo.Term)
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("登陆模块 SMS 平台配置有误"))
		}
	}
}

// SSOSignUp 单点登模块注册接口的处理函数
func SSOSignUp(c *gin.Context) {
	// 构造会话信息
	sessionInfo := structs.SingleSignOnSession{}
	// 绑定数据
	if err := c.ShouldBindJSON(&sessionInfo); err != nil {
		// 绑定数据错误
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}
	sessionInfo.SourceIP = c.ClientIP()

	// 验证手机号码是否有效
	if !verify.Phone(sessionInfo.Phone) {
		c.JSON(http.StatusBadRequest, response.Phone.Invalid)
		return
	}

	// 校验登录模块配置
	moduleInfo := structs.SingleSignOnLoginModule{}
	orm.MySQL.Gaea.Where("id = ?", sessionInfo.MID).Find(&moduleInfo)
	if moduleInfo.ID == 0 {
		// 模块不存在
		c.JSON(http.StatusBadRequest, response.Message("单点登陆模块配置有误"))
		return
	}
	// 保存模块信息到会话信息中
	sessionInfo.CRMEFSID = moduleInfo.CRMEFSID // CRM 活动表单 SID

	// 检查验证码是否正确且未失效
	codeInfo := structs.SingleSignOnVerificationCode{}
	orm.MySQL.Gaea.Where("phone = ?", sessionInfo.Phone).Find(&codeInfo)
	// 校验是否发送过验证码
	if codeInfo.ID == 0 {
		c.JSON(http.StatusBadRequest, response.Message("请您先获取验证码后再进行注册"))
		return
	}
	// 校验验证码是正确
	if sessionInfo.Code != codeInfo.Code {
		c.JSON(http.StatusBadRequest, response.Message("验证码有误"))
		return
	}
	// 校验验证码是否有效
	duration, _ := time.ParseDuration("-" + fmt.Sprint(codeInfo.Term) + "m")
	if codeInfo.CreatedAt.Before(time.Now().Add(duration)) {
		c.JSON(http.StatusBadRequest, response.Message("验证码失效"))
		return
	}

	// 校验用户是否已经注册 ( 避免重复注册 )
	if !ssoIsSignUp(sessionInfo.Phone) {
		// 保存注册信息
		orm.MySQL.Gaea.Create(&structs.SingleSignOnUser{Phone: sessionInfo.Phone})
	}

	// 校验用户是否已经参与过当前活动 ( 避免重复创建会话信息, 避免重复推送信息到 CRM )
	if !ssoIsSignIn(sessionInfo.Phone, sessionInfo.MID) {
		// 用户未参与过当前活动, 保存会话
		ssoCreateSession(&sessionInfo)
	}

	// 注册成功
	c.JSON(http.StatusOK, response.Success)
}

// SSOSignIn 单点登陆模块登陆接口的处理函数
func SSOSignIn(c *gin.Context) {
	// 构造会话信息
	sessionInfo := structs.SingleSignOnSession{}
	// 绑定数据
	if err := c.ShouldBindJSON(&sessionInfo); err != nil {
		// 绑定数据错误
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}
	sessionInfo.SourceIP = c.ClientIP()

	// 验证手机号码是否有效
	if !verify.Phone(sessionInfo.Phone) {
		c.JSON(http.StatusBadRequest, response.Phone.Invalid)
		return
	}

	// 校验登录模块配置
	moduleInfo := structs.SingleSignOnLoginModule{}
	orm.MySQL.Gaea.Where("id = ?", sessionInfo.MID).Find(&moduleInfo)
	if moduleInfo.ID == 0 {
		// 模块不存在
		c.JSON(http.StatusBadRequest, response.Message("单点登陆模块配置有误"))
		return
	}
	// 保存 CRM 活动表单 ID 到会话信息中
	sessionInfo.CRMEFSID = moduleInfo.CRMEFSID // CRM 活动表单 ID

	// 校验用户是否已经注册 ( 避免重复注册 )
	if !ssoIsSignUp(sessionInfo.Phone) {
		// 保存注册信息
		c.JSON(http.StatusForbidden, response.Message("请您先进行注册"))
		return
	}

	// 校验用户是否已经参与过当前活动 ( 避免重复创建会话信息, 避免重复推送信息到 CRM )
	if !ssoIsSignIn(sessionInfo.Phone, sessionInfo.MID) {
		// 未参与
		// 保存会话
		ssoCreateSession(&sessionInfo)
	}

	// 登陆成功
	c.JSON(http.StatusOK, response.Success)
}

// SSOSessionInfo 获取会话信息
func SSOSessionInfo(c *gin.Context) {
	responseData := struct {
		Sign           string // 发信签名
		CRMEID         string // CRM 活动 ID
		CRMEFID        uint   // CRM 活动表单 ID
		CRMEFSID       string // CRM 活动表单 SID
		CRMChannel     uint   // CRM 所属渠道
		CRMOCode       uint   // CRM 组织代码
		CRMOName       string // CRM 组织名称
		CRMUID         uint   // CRM 用户 ID
		CRMUser        string // CRM 用户名
		Suffix         string // 后缀 ( 19课堂后缀 )
		NTalkerGID     string // 小能咨询组
		IsLogin        bool   // 是否登陆
		NeedToRegister bool   // 是否需要注册
	}{}

	// 验证手机号是否有效
	if c.Param("Phone") != "0" && !verify.Phone(c.Param("Phone")) {
		c.JSON(http.StatusBadRequest, response.Phone.Invalid)
		return
	}

	// 校验登陆模块配置
	moduleInfo := structs.SingleSignOnLoginModule{}
	orm.MySQL.Gaea.Where("id = ?", c.Param("MID")).Find(&moduleInfo)
	if moduleInfo.ID == 0 {
		// 模块不存在
		c.JSON(http.StatusBadRequest, response.Message("单点登陆模块配置有误!"))
		return
	}
	// 保存模块信息到会话信息中
	responseData.Sign = moduleInfo.Sign         // 发信签名
	responseData.CRMEID = moduleInfo.CRMEID     // CRM 活动 ID
	responseData.CRMEFID = moduleInfo.CRMEFID   // CRM 活动表单 ID
	responseData.CRMEFSID = moduleInfo.CRMEFSID // CRM 活动表单 SID

	// 校验后缀
	suffixInfo := structs.SingleSignOnSuffix{}
	orm.MySQL.Gaea.Unscoped().Where("suffix = ?", c.Param("Suffix")).Find(&suffixInfo)
	if suffixInfo.ID == 0 {
		// 后缀不存在
		// 获取默认后缀 ( ID = 1, 第一条 )
		defaultSuffixInfo := structs.SingleSignOnSuffix{}
		orm.MySQL.Gaea.First(&defaultSuffixInfo)
		responseData.CRMChannel = defaultSuffixInfo.CRMChannel // CRM 所属渠道
		responseData.CRMUID = defaultSuffixInfo.CRMUID         // CRM 用户 ID
		responseData.CRMUser = defaultSuffixInfo.CRMUser       // CRM 用户名
		responseData.Suffix = defaultSuffixInfo.Suffix         // 后缀 ( 19课堂后缀 )
		responseData.NTalkerGID = defaultSuffixInfo.NTalkerGID // 小能咨询组
		// 获取默认后缀对应的 CRM 组织信息
		organizationInfo := structs.SingleSignOnOrganization{}
		if defaultSuffixInfo.CRMOID == 0 {
			// 当前后缀未配置归属组织, 获取省级分部的信息
			orm.MySQL.Gaea.Where("f_id = 0").Find(&organizationInfo)
		} else {
			// 获取当前后缀配置的归属分部信息
			orm.MySQL.Gaea.Where("id = ?", defaultSuffixInfo.CRMOID).Find(&organizationInfo)
			if organizationInfo.ID == 0 {
				// 获取当前后缀配置的归属分部信息失败, 获取省级分部信息
				orm.MySQL.Gaea.Where("f_id = 0").Find(&organizationInfo)
			}
		}
		responseData.CRMOCode = organizationInfo.Code
		responseData.CRMOName = organizationInfo.Name
	} else {
		// 后缀存在
		responseData.CRMChannel = suffixInfo.CRMChannel // CRM 所属渠道
		responseData.CRMUID = suffixInfo.CRMUID         // CRM 用户 ID
		responseData.CRMUser = suffixInfo.CRMUser       // CRM 用户名
		responseData.Suffix = suffixInfo.Suffix         // 后缀 ( 19课堂后缀 )
		responseData.NTalkerGID = suffixInfo.NTalkerGID // 小能咨询组
		// 获取 CRM 组织信息
		organizationInfo := structs.SingleSignOnOrganization{}
		if suffixInfo.CRMOID == 0 {
			// 当前后缀未配置归属组织, 获取省级分部的信息
			orm.MySQL.Gaea.Where("f_id = 0").Find(&organizationInfo)
		} else {
			// 获取当前后缀配置的归属分部信息
			orm.MySQL.Gaea.Where("id = ?", suffixInfo.CRMOID).Find(&organizationInfo)
			if organizationInfo.ID == 0 {
				// 获取当前后缀配置的归属分部信息失败, 获取省级分部信息
				orm.MySQL.Gaea.Where("f_id = 0").Find(&organizationInfo)
			}
		}
		responseData.CRMOCode = organizationInfo.Code
		responseData.CRMOName = organizationInfo.Name
	}

	// 校验是否需要注册
	userInfo := structs.SingleSignOnUser{}
	orm.MySQL.Gaea.Where("phone = ? and created_at > ?", c.Param("Phone"), time.Now().AddDate(0, 0, -30)).Find(&userInfo)
	if userInfo.ID == 0 {
		// 未进行注册, 需要注册
		responseData.NeedToRegister = true
	}

	// 校验是否需要登陆
	sessionInfo := structs.SingleSignOnSession{}
	orm.MySQL.Gaea.Where("phone = ? and m_id = ?", c.Param("Phone"), moduleInfo.ID).Find(&sessionInfo)
	if sessionInfo.ID != 0 {
		// 已经登陆
		responseData.IsLogin = true
	}

	// 返回会话信息
	c.JSON(http.StatusOK, response.Data(responseData))
}

// sendVerificationCodeByOFFCN 使用中公短信平台发送验证码
// 签名无需设置亦无法变更, 所以忽略签名参数
// 平台没有模板的概念, 但是为了更加通用, 内部模拟一套与腾讯云短信服务相同的模板逻辑, 基于格式化输出实现
func sendVerificationCodeByOFFCN(c *gin.Context, templateID, term uint) {
	// 验证模板 ID 并配置模板内容
	template := ""
	switch templateID {
	case 391863:
		// 验证码 ( 登陆 )
		template = "%d 为您的登录验证码，请于 %d 分钟内填写。如非本人操作，请忽略本短信。"
	case 392030:
		// 通用验证码
		template = "您的验证码是 %d ，请于 %d 分钟内填写。如非本人操作，请忽略本短信。"
	case 392074:
		// 可复用验证码
		template = "您的验证码是 %d ，%d 分钟内可重复使用。如非本人操作，请忽略本短信。"
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("短信模板配置有误"))
		return
	}

	// 检查配置
	if config.Get().OffcnSmsURL == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("未配置 中公教育短信平台 接口地址"))
		return
	}
	if config.Get().OffcnSmsUserName == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("未配置 中公教育短信平台 用户名"))
		return
	}
	if config.Get().OffcnSmsPassword == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("未配置 中公教育短信平台 密码"))
		return
	}
	if config.Get().OffcnSmsTjCode == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("未配置 中公教育短信平台 发送方识别码"))
		return
	}

	// 验证是否具有发送条件
	verificationCodeInfo := structs.SingleSignOnVerificationCode{}
	orm.MySQL.Gaea.Where("phone = ?", c.Param("Phone")).Find(&verificationCodeInfo)
	if verificationCodeInfo.ID != 0 {
		// 存在发送记录, 继续判断是否失效
		duration, _ := time.ParseDuration("-" + fmt.Sprint(verificationCodeInfo.Term) + "m")
		if verificationCodeInfo.CreatedAt.After(time.Now().Add(duration)) {
			// 上一条验证码未超过有效期
			c.AbortWithStatusJSON(http.StatusBadRequest, response.Message("请勿重复发送验证码"))
			return
		}
	}

	// 初始化随机数的资源库, 如果不执行这行, 不管运行多少次都返回同样的值 # https://learnku.com/articles/26011
	rand.Seed(time.Now().UnixNano())
	// 生成随机数作为验证码
	code := uint(rand.Intn(8999) + 1000) // 如果直接用 Intn(9999) 会有可能生成出来不是4位的数字 ( 小于 1000 的数字 )

	// 拼接参数
	data := url.Values{
		"sname":   []string{config.Get().OffcnSmsUserName},
		"spwd":    []string{config.Get().OffcnSmsPassword},
		"mobile":  []string{c.Param("Phone")},
		"content": []string{fmt.Sprintf(template, code, term)},
		"tjcode":  []string{config.Get().OffcnSmsTjCode},
	}

	// 发送短信
	if resp, err := http.PostForm(config.Get().OffcnSmsURL, data); err != nil {
		logger.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("发送短信失败, 返回错误 : "+err.Error()))
	} else {
		defer resp.Body.Close()
		// 判断有没有发送成功
		if resp.StatusCode != 200 {
			// 请求出错
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("发送短信失败, 返回状态码 : "+fmt.Sprint(resp.StatusCode)))
		} else {
			// 读取 body
			if respBytes, err := ioutil.ReadAll(resp.Body); err != nil {
				logger.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("发送短信失败, 读取返回内容失败, 错误内容 : "+err.Error()))
			} else {
				// 解码 body
				var respJsonMap map[string]interface{}
				if err := json.Unmarshal(respBytes, &respJsonMap); err != nil {
					logger.Error(err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("发送短信失败, 解码返回内容失败, 错误内容 : "+err.Error()))
				} else {
					// 返回请求回来的 Json 的 Map
					if respJsonMap["status"].(float64) != 1 {
						// 发送失败
						c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("发送短信失败, [ "+fmt.Sprint(respJsonMap["status"])+" ] "+fmt.Sprint(respJsonMap["msg"])))
					} else {
						// 发送成功
						// 保存验证码发送记录
						orm.MySQL.Gaea.Create(&structs.SingleSignOnVerificationCode{Phone: c.Param("Phone"), Term: term, Code: code, SourceIP: c.ClientIP()})
						c.JSON(http.StatusOK, response.Success)
					}
				}
			}
		}
	}
}

// sendVerificationCodeByTencentCloudSMSV2 使用腾讯云短信平台发送验证码
func sendVerificationCodeByTencentCloudSMSV2(c *gin.Context, sign string, templateID, term uint) {
	// 检查配置
	if config.Get().TencentCloudAPISecretID == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("未配置 腾讯云 令牌"))
		return
	}
	if config.Get().TencentCloudAPISecretKey == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("未配置 腾讯云 密钥"))
		return
	}
	if config.Get().TencentCloudSmsSdkAppId == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("未配置 腾讯云 短信应用 ID"))
		return
	}

	// 初始化随机数的资源库, 如果不执行这行, 不管运行多少次都返回同样的值 # https://learnku.com/articles/26011
	rand.Seed(time.Now().UnixNano())
	// 生成随机数作为验证码
	code := uint(rand.Intn(8999) + 1000) // 如果直接用 Intn(9999) 会生成出来不是4位的数字

	// 验证是否具有发送条件
	verificationCodeInfo := structs.SingleSignOnVerificationCode{}
	orm.MySQL.Gaea.Where("phone = ?", c.Param("Phone")).Find(&verificationCodeInfo)
	if verificationCodeInfo.ID != 0 {
		// 存在发送记录, 继续判断是否失效
		duration, _ := time.ParseDuration("-" + fmt.Sprint(verificationCodeInfo.Term) + "m")
		if verificationCodeInfo.CreatedAt.After(time.Now().Add(duration)) {
			// 上一条验证码未超过有效期
			c.AbortWithStatusJSON(http.StatusBadRequest, response.Message("请勿重复发送验证码"))
			return
		}
	}

	// # https://cloud.tencent.com/document/product/382/43199
	/* 必要步骤：
	 * 实例化一个认证对象，入参需要传入腾讯云账户密钥对 secretId 和 secretKey
	 * 本示例采用从环境变量读取的方式，需要预先在环境变量中设置这两个值
	 * 您也可以直接在代码中写入密钥对，但需谨防泄露，不要将代码复制、上传或者分享给他人
	 * CAM 密匙查询: https://console.cloud.tencent.com/cam/capi*/
	credential := common.NewCredential(config.Get().TencentCloudAPISecretID, config.Get().TencentCloudAPISecretKey)

	/* 非必要步骤:
	 * 实例化一个客户端配置对象，可以指定超时时间等配置 */
	cpf := profile.NewClientProfile()

	/* SDK 默认使用 POST 方法
	 * 如需使用 GET 方法，可以在此处设置，但 GET 方法无法处理较大的请求 */
	cpf.HttpProfile.ReqMethod = "POST"

	/* SDK 有默认的超时时间，非必要请不要进行调整
	 * 如有需要请在代码中查阅以获取最新的默认值 */
	//cpf.HttpProfile.ReqTimeout = 5

	/* SDK 会自动指定域名，通常无需指定域名，但访问金融区的服务时必须手动指定域名
	 * 例如 SMS 的上海金融区域名为 sms.ap-shanghai-fsi.tencentcloudapi.com */
	//cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	cpf.HttpProfile.Endpoint = "sms.internal.tencentcloudapi.com" // 使用内网接口地址

	/* SDK 默认用 TC3-HMAC-SHA256 进行签名，非必要请不要修改该字段 */
	cpf.SignMethod = "HmacSHA1"

	/* 实例化 SMS 的 client 对象
	 * 第二个参数是地域信息，可以直接填写字符串 ap-guangzhou，或者引用预设的常量 */
	//client, _ := sms.NewClient(credential, "ap-guangzhou", cpf)
	client, _ := sms.NewClient(credential, regions.Beijing, cpf)

	/* 实例化一个请求对象，根据调用的接口和实际情况，可以进一步设置请求参数
	   * 您可以直接查询 SDK 源码确定接口有哪些属性可以设置
	    * 属性可能是基本类型，也可能引用了另一个数据结构
	    * 推荐使用 IDE 进行开发，可以方便地跳转查阅各个接口和数据结构的文档说明 */
	request := sms.NewSendSmsRequest()

	/* 基本类型的设置:
	 * SDK 采用的是指针风格指定参数，即使对于基本类型也需要用指针来对参数赋值。
	 * SDK 提供对基本类型的指针引用封装函数
	 * 帮助链接：
	 * 短信控制台：https://console.cloud.tencent.com/smsv2
	 * sms helper：https://cloud.tencent.com/document/product/382/3773 */

	/* 短信应用 ID: 在 [短信控制台] 添加应用后生成的实际 SDKAppID，例如1400006666 */
	request.SmsSdkAppid = common.StringPtr(config.Get().TencentCloudSmsSdkAppId)
	/* 短信签名内容: 使用 UTF-8 编码，必须填写已审核通过的签名，可登录 [短信控制台] 查看签名信息 */
	request.Sign = common.StringPtr(sign)
	/* 国际/港澳台短信 senderid: 国内短信填空，默认未开通，如需开通请联系 [sms helper] */
	//request.SenderId = common.StringPtr("xxx")
	/* 用户的 session 内容: 可以携带用户侧 ID 等上下文信息，server 会原样返回 */
	//request.SessionContext = common.StringPtr("xxx")
	/* 短信码号扩展号: 默认未开通，如需开通请联系 [sms helper] */
	//request.ExtendCode = common.StringPtr("0")
	/* 模板参数: 若无模板参数，则设置为空*/
	request.TemplateParamSet = common.StringPtrs([]string{fmt.Sprint(code), fmt.Sprint(term)})
	/* 模板 ID: 必须填写已审核通过的模板 ID，可登录 [短信控制台] 查看模板 ID */
	request.TemplateID = common.StringPtr(fmt.Sprint(templateID))
	/* 下发手机号码，采用 e.164 标准，+[国家或地区码][手机号]
	 * 例如+8613711112222， 其中前面有一个+号 ，86为国家码，13711112222为手机号，最多不要超过200个手机号*/
	request.PhoneNumberSet = common.StringPtrs([]string{"+86" + c.Param("Phone")})

	// 通过 client 对象调用想要访问的接口，需要传入请求对象
	responseData, err := client.SendSms(request)
	// 处理异常
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		logger.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("发送短信失败, [ "+err.(*errors.TencentCloudSDKError).GetCode()+" ] "+err.(*errors.TencentCloudSDKError).GetMessage()))
		return
	}
	// 非 SDK 异常，直接失败。实际代码中可以加入其他的处理
	if err != nil {
		logger.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Message("发送短信失败, 未知错误"))
		return
	}

	if *responseData.Response.SendStatusSet[0].Code != "Ok" {
		logger.DebugToJson("腾讯云短信平台响应内容", responseData.Response)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Message("发送短信失败, 错误内容 : "+*responseData.Response.SendStatusSet[0].Message))
		return
	}

	// 发送成功
	// 保存验证码发送记录
	orm.MySQL.Gaea.Create(&structs.SingleSignOnVerificationCode{Phone: c.Param("Phone"), Term: term, Code: code, SourceIP: c.ClientIP()})
	c.JSON(http.StatusOK, response.Success)
}

// ssoIsSignUp 内部函数 检查用户是否已经注册且未失效
func ssoIsSignUp(phone string) bool {
	userInfo := structs.SingleSignOnUser{}
	orm.MySQL.Gaea.Where("phone = ? and created_at > ?", phone, time.Now().AddDate(0, 0, -30)).Find(&userInfo)
	if userInfo.ID != 0 {
		return true
	}
	return false
}

// ssoIsSignIn 内部函数 检查用户是否已经登陆
func ssoIsSignIn(phone string, mID uint) bool {
	sessionInfo := structs.SingleSignOnSession{}
	orm.MySQL.Gaea.Where("phone = ? and m_id = ?", phone, mID).Find(&sessionInfo)
	if sessionInfo.ID != 0 {
		return true
	}
	return false
}

// ssoCreateSession 内部函数 按照预期矫正信息后创建会话
// 创建会话前会进行推送信息到 CRM 的操作
func ssoCreateSession(session *structs.SingleSignOnSession) {
	// 校验后缀
	if session.ActualSuffix == "" {
		// 后缀未填写, 使用默认后缀配置
		SSOGetDefaultSuffix(session)
	} else {
		suffixInfo := structs.SingleSignOnSuffix{}
		orm.MySQL.Gaea.Unscoped().Where("suffix = ?", session.ActualSuffix).Find(&suffixInfo)
		if suffixInfo.ID == 0 {
			// 后缀无效, 使用默认后缀配置
			SSOGetDefaultSuffix(session)
		} else {
			session.CRMChannel = suffixInfo.CRMChannel
			session.CRMUID = suffixInfo.CRMUID
			session.CurrentSuffix = suffixInfo.Suffix
			if suffixInfo.CRMOID > 1 {
				// 配置了 CRMOID 并且不是省级
				organizationInfo := structs.SingleSignOnOrganization{}
				orm.MySQL.Gaea.Where("id = ?", suffixInfo.CRMOID).Find(&organizationInfo)
				session.CRMOCode = organizationInfo.Code
			} else {
				// 未配置 CRMOID 或者是省级 ( 等于 1 ), 按手机号码归属地分配 CRM 信息
				SSODistributionByPhoneNumber(session)
			}
		}
	}

	// 推送信息到 CRM
	urlObject, _ := url.Parse("https://dc.offcn.com:8443/a.gif") // 此处为固定链接, 在解析 URL 的过程中不可能出现错误, 所以对返回对 err 进行忽略
	// 构建参数 queryObject
	queryObject := urlObject.Query()
	queryObject.Set("sid", session.CRMEFSID)
	queryObject.Set("mobile", session.Phone)
	queryObject.Set("channel", fmt.Sprint(session.CRMChannel))
	queryObject.Set("orgn", fmt.Sprint(session.CRMOCode))
	if session.CRMUID != 0 {
		queryObject.Set("owner", fmt.Sprint(session.CRMUID))
	}
	if session.CustomerName != "" {
		queryObject.Set("name", session.CustomerName)
	}
	if session.CustomerIdentityID != 0 {
		queryObject.Set("khsf", fmt.Sprint(session.CustomerIdentityID))
	}
	if session.CustomerColleage != "" {
		queryObject.Set("colleage", session.CustomerColleage)
	}
	if session.CustomerMayor != "" {
		queryObject.Set("mayor", session.CustomerMayor)
	}
	if session.Remark != "" {
		queryObject.Set("remark", session.Remark)
	}
	// 发送 GET 请求
	urlObject.RawQuery = queryObject.Encode()
	if getResponse, err := http.Get(urlObject.String()); err != nil {
		// 发送 GET 请求出错
		logger.Error(err)
		// 推送失败, 保存推送失败记录
		orm.MySQL.Gaea.Create(&structs.SingleSignOnErrorLog{
			Phone:      session.Phone,
			MID:        session.MID,
			CRMChannel: session.CRMChannel,
			CRMUID:     session.CRMUID,
			CRMOCode:   session.CRMOCode,
			Error:      err.Error(),
		})
	} else {
		if getResponse.StatusCode != 200 {
			// 推送失败, 保存推送失败记录
			orm.MySQL.Gaea.Create(&structs.SingleSignOnErrorLog{
				Phone:      session.Phone,
				MID:        session.MID,
				CRMChannel: session.CRMChannel,
				CRMUID:     session.CRMUID,
				CRMOCode:   session.CRMOCode,
				Error:      "CRM 响应状态码 : " + fmt.Sprint(getResponse.StatusCode),
			})
		}
	}

	// 保存会话信息
	orm.MySQL.Gaea.Create(&session)
}

// SSOGetDefaultSuffix 内部函数 获取默认后缀配置
func SSOGetDefaultSuffix(session *structs.SingleSignOnSession) {
	session.CRMChannel = 7 // 默认所属渠道 19课堂
	// 获取默认后缀
	suffixInfo := structs.SingleSignOnSuffix{}
	orm.MySQL.Gaea.First(&suffixInfo)
	// 配置默认后缀
	session.CurrentSuffix = suffixInfo.Suffix
	session.CRMUID = suffixInfo.CRMUID
	// 所属组织按照手机号码归属地进行分配
	SSODistributionByPhoneNumber(session)
}

// SSODistributionByPhoneNumber 按照手机号码归属地进行归属分部分配
func SSODistributionByPhoneNumber(session *structs.SingleSignOnSession) {
	if record, err := phonedata.Find(session.Phone); err != nil {
		logger.Error(err)
		// 解析出错，循环分配给九个地市
		ssoRoundCrmList(session)
	} else {
		switch record.City {
		case "长春":
			session.CRMOCode = 2290
		case "吉林":
			session.CRMOCode = 2305
		case "延边":
			session.CRMOCode = 2277
		case "通化":
			session.CRMOCode = 2271
		case "白山":
			session.CRMOCode = 2310
		case "四平":
			session.CRMOCode = 2263
		case "松原":
			session.CRMOCode = 2284
		case "白城":
			session.CRMOCode = 2315
		case "辽源":
			session.CRMOCode = 2268
		default:
			// 循环分配给九个地市
			ssoRoundCrmList(session)
		}
	}
}

// ssoRoundCrmList 内部函数 循环分配手机号给九个地市分部
// 高并发时存在数据库读写延迟，可能无法确保幂等性
// 即, 可能出现同时某分部重复分配, 或跳过某分部进行分配的情况
func ssoRoundCrmList(session *structs.SingleSignOnSession) {
	// 取出地市分校列表
	crmOrganizations := make([]structs.SingleSignOnOrganization, 0)
	orm.MySQL.Gaea.Where("f_id = 1").Find(&crmOrganizations)
	logger.DebugToJson("crmOrganizations", crmOrganizations)

	// 获取当前分配计数
	count := 0
	orm.MySQL.Gaea.Model(structs.SingleSignOnCRMRoundLog{}).Count(&count)
	logger.DebugToString("count", count)

	// 分配
	session.CRMOCode = crmOrganizations[count%len(crmOrganizations)].Code

	// 保存分配记录
	orm.MySQL.Gaea.Create(&structs.SingleSignOnCRMRoundLog{
		MID:    session.MID,
		Phone:  session.Phone,
		CRMOID: crmOrganizations[count%len(crmOrganizations)].ID,
	})
}
