/*
   @Time : 2020/11/29 10:04 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : request_test
   @Software: GoLand
   @Description: 请求工具的单元测试
*/

package request

import (
	"fmt"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

// TestGetSendQueryReceiveBytes 测试 GetSendQueryReceiveBytes 是否可以用 GET 发送 QueryString 类型的请求并接受 Bytes 类型的响应
func TestGetSendQueryReceiveBytes(t *testing.T) {
	Convey("测试 GetSendQueryReceiveBytes 是否可以用 GET 发送 QueryString 类型的请求并接受 Bytes 类型的响应", t, func() {
		// 测试解析请求失败
		responseBytes, err := GetSendQueryReceiveBytes("\n", nil)
		So(responseBytes, ShouldBeNil)
		So(err.Error(), ShouldEqual, "parse \"\\n\": net/url: invalid control character in URL")

		// 测试发送请求失败
		responseBytes, err = GetSendQueryReceiveBytes("", nil)
		So(responseBytes, ShouldBeNil)
		So(err.Error(), ShouldEqual, "Get \"\": unsupported protocol scheme \"\"")

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://fake.request", httpmock.NewBytesResponder(http.StatusInternalServerError, nil))

		// 测试返回错误的状态码
		responseBytes, err = GetSendQueryReceiveBytes("https://fake.request", nil)
		So(responseBytes, ShouldBeNil)
		So(err.Error(), ShouldEqual, "发送 GET 请求出错. 状态码: 500")

		// 调整 httpmock 返回正确的响应
		httpmock.RegisterResponder(http.MethodGet, "https://fake.request", func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(http.StatusOK, req.URL.String()), nil
		})

		// 初始化 ORM

		// 测试返回正确的数据
		responseBytes, err = GetSendQueryReceiveBytes("https://fake.request", map[string]string{"foo": "bar"})
		So(string(responseBytes), ShouldEqual, "https://fake.request?foo=bar")
		So(err, ShouldBeNil)
	})
}

// TestGetSendQueryReceiveJson 测试 GetSendQueryReceiveJson 是否可以用 GET 发送 QueryString 类型的请求并接受 Json 类型的响应
func TestGetSendQueryReceiveJson(t *testing.T) {
	Convey("测试 GetSendQueryReceiveJson 是否可以用 GET 发送 QueryString 类型的请求并接受 Json 类型的响应 fixme 单元测试", t, func() {
		responseJson, err := GetSendQueryReceiveJson("", nil)
		So(responseJson, ShouldBeNil)
		So(err.Error(), ShouldEqual, "Get \"\": unsupported protocol scheme \"\"")

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://fake.request", httpmock.NewBytesResponder(http.StatusOK, nil))

		// 测试返回错误的 Json
		responseJson, err = GetSendQueryReceiveJson("https://fake.request", nil)
		So(responseJson, ShouldBeNil)
		So(err.Error(), ShouldEqual, "unexpected end of JSON input")

		// 调整 httpmock 返回正确的响应
		httpmock.RegisterResponder(http.MethodGet, "https://fake.request", func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(http.StatusOK, req.URL.Query())
		})

		// 测试返回正确的数据
		responseJson, err = GetSendQueryReceiveJson("https://fake.request", map[string]string{"foo": "bar"})
		So(fmt.Sprint(responseJson), ShouldEqual, "map[foo:[bar]]")
		So(err, ShouldBeNil)
	})
}
