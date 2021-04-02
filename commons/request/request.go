/*
   @Time : 2020/11/28 5:06 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : request
   @Software: GoLand
   @Description: 请求工具
*/

package request

import (
	"bytes"
	"encoding/json"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"io/ioutil"
	"net/http"
	"net/url"
)

// GetSendQueryReceiveBytes 用 GET 发送 QueryString 类型的请求并接受 Bytes 类型的响应
func GetSendQueryReceiveBytes(path string, query map[string]string) ([]byte, error) {
	// 解析请求
	if urlObject, err := url.Parse(path); err != nil {
		return nil, err
	} else {
		// 将参数 map 拼接到 query 字符串中
		queryObject := urlObject.Query()
		for queryKey, queryValue := range query {
			queryObject.Set(queryKey, queryValue)
		}
		urlObject.RawQuery = queryObject.Encode()
		// 发送 GET 请求
		if responseData, err := http.Get(urlObject.String()); err != nil {
			// 发送 GET 请求出错
			return nil, err
		} else {
			defer responseData.Body.Close() // 函数退出时关闭 body
			// 读取 body
			if responseBytes, err := ioutil.ReadAll(responseData.Body); err != nil {
				return nil, err
			} else {
				logger.DebugToString("GET 请求到 "+path+" 的响应", string(responseBytes))
				return responseBytes, nil
			}
		}
	}
}

// GetSendQueryReceiveJson 用 GET 发送 QueryString 类型的请求并接受 Json 类型的响应
func GetSendQueryReceiveJson(path string, query map[string]string) (map[string]interface{}, error) {
	if responseBytes, err := GetSendQueryReceiveBytes(path, query); err != nil {
		return nil, err
	} else {
		// 解码 responseBytes
		var responseJson map[string]interface{}
		if err := json.Unmarshal(responseBytes, &responseJson); err != nil {
			return nil, err
		} else {
			// 返回请求回来的 Json 的 Map
			return responseJson, nil
		}
	}
}

// PostSendJsonReceiveBytes 用 POST 发送 Json 类型的请求并接受 Bytes 类型响应
func PostSendJsonReceiveBytes(path string, jsonMap map[string]interface{}) ([]byte, error) {
	// 将 jsonMap 序列化为 Json 字符串并不进行转义
	// 因为微信公众平台的创建小程序二维码接口不能接收 htmlEncode 编码过的 Json 数据, 所以采用这种方式生成 Json 字符串
	requestBody := &bytes.Buffer{}
	encoder := json.NewEncoder(requestBody)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(jsonMap); err != nil {
		return nil, err
	} else {
		logger.DebugToString("序列化后的 Json 字符串", requestBody.String())
		if responseData, err := http.Post(path, "application/json", requestBody); err != nil {
			return nil, err
		} else {
			defer responseData.Body.Close() // 函数退出时关闭 body
			// 读取 body
			if responseBytes, err := ioutil.ReadAll(responseData.Body); err != nil {
				return nil, err
			} else {
				logger.DebugToJson("POST 请求到 "+path+" 的响应", string(responseBytes))
				return responseBytes, nil
			}
		}
	}
}
