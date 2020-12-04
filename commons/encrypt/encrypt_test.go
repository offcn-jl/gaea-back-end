/*
   @Time : 2020/12/3 5:52 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : encrypt_test
   @Software: GoLand
   @Description: 加密工具的单元测试
*/

package encrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	privateKeyString, publicKeyString string
)

func init() {
	// 生成私钥
	// # https://www.cnblogs.com/PeterXu1997/p/12218553.html
	// # https://blog.csdn.net/chenxing1230/article/details/83757638
	// 生成 RSA 密钥对
	// GenerateKey 函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	// Reader 是一个全局、共享的密码用强随机数生成器
	privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	// 通过 x509 标准将得到的 ras 私钥序列化为 ASN.1 的 DER 编码字符串
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	// 构建一个 pem.Block 结构体对象
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	// 初始化用于接收 pem 的 buffer
	bufferPrivate := new(bytes.Buffer)
	// 使用 pem 格式对 x509 输出的内容进行编码
	pem.Encode(bufferPrivate, &privateBlock)
	privateKeyString = bufferPrivate.String()

	// 生成公钥
	// X509 对公钥编码
	X509PublicKey, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	//创建一个pem.Block结构体对象
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	// 初始化用于接收 pem 的 buffer
	bufferPublic := new(bytes.Buffer)
	// pem格式编码
	pem.Encode(bufferPublic, &publicBlock)
	publicKeyString = bufferPublic.String()

	// 初始化数据库
	utt.InitTest() // 初始化测试数据并获取测试所需的上下文
}

// TestRSADecrypt 测试 RSADecrypt 是否可以进行 RSA 解密
func TestRSADecrypt(t *testing.T) {
	Convey("测试 RSADecrypt 是否可以进行 RSA 解密", t, func() {
		// 测试需要解密的字符串 Base64 解码失败
		decryptedString, err := RSADecrypt("*")
		So(decryptedString, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "illegal base64 data at input byte 0")

		// 测试 RSA 私钥 PEM 解码失败
		decryptedString, err = RSADecrypt("")
		So(decryptedString, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "RSA 私钥 PEM 解码失败")

		// 配置错误的私钥 ( 此处将公钥配置为私钥 )
		config.Update(utt.ORM, structs.SystemConfig{RSAPrivateKey: publicKeyString})

		// 测试 私钥 X509 解码失败
		decryptedString, err = RSADecrypt("")
		So(decryptedString, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "asn1: structure error: tags don't match (2 vs {class:0 tag:16 length:13 isCompound:true}) {optional:false explicit:false application:false private:false defaultValue:<nil> tag:<nil> stringType:0 timeType:0 set:false omitEmpty:false} int @2")

		// 配置正确的密钥对
		config.Update(utt.ORM, structs.SystemConfig{RSAPublicKey: publicKeyString, RSAPrivateKey: privateKeyString})

		// 测试解密成功
		encryptedString, _ := RSAEncrypt([]byte("foobar"))
		decryptedString, err = RSADecrypt(encryptedString)
		So(string(decryptedString), ShouldEqual, "foobar")
		So(err, ShouldBeNil)

		// 移除配置, 避免干扰后续测试
		config.Update(utt.ORM, structs.SystemConfig{})
	})
}

// TestRSAEncrypt 测试 RSAEncrypt 是否可以进行 RSA 加密
func TestRSAEncrypt(t *testing.T) {
	Convey("测试 RSAEncrypt 是否可以进行 RSA 加密", t, func() {
		// 测试 RSA 公钥 PEM 解码失败
		encryptedString, err := RSAEncrypt([]byte("test-data"))
		So(encryptedString, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "RSA 公钥 PEM 解码失败")

		// 配置错误的公钥 ( 此处将私钥配置为公钥 )
		config.Update(utt.ORM, structs.SystemConfig{RSAPublicKey: privateKeyString})

		// 测试公钥 X509 解码失败
		encryptedString, err = RSAEncrypt([]byte("test-data"))
		So(encryptedString, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "asn1: structure error: tags don't match (16 vs {class:0 tag:2 length:1 isCompound:false}) {optional:false explicit:false application:false private:false defaultValue:<nil> tag:<nil> stringType:0 timeType:0 set:false omitEmpty:false} AlgorithmIdentifier @2")

		// 配置正确的公钥
		config.Update(utt.ORM, structs.SystemConfig{RSAPublicKey: publicKeyString})

		// 测试加密失败
		encryptedString, err = RSAEncrypt([]byte(publicKeyString + publicKeyString))
		So(encryptedString, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "crypto/rsa: message too long for RSA public key size")

		// 测试加密成功
		encryptedString, err = RSAEncrypt([]byte("foobar"))
		So(encryptedString, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
