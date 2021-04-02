/*
   @Time : 2020/12/3 10:11 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : encrypt
   @Software: GoLand
   @Description: 加密工具
*/

package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"github.com/offcn-jl/gaea-back-end/commons/config"
)

// RSADecrypt RSA 解密
func RSADecrypt(ciphertext string) ([]byte, error) {
	if decodedString, err := base64.StdEncoding.DecodeString(ciphertext); err != nil {
		// 字符串 BASE64 解码失败
		return nil, err
	} else {
		// 解码成功, 进行 RSA 解密
		// pem 解码
		privateKeyBlock, _ := pem.Decode([]byte(config.Get().RSAPrivateKey))
		if privateKeyBlock == nil {
			return nil, errors.New("RSA 私钥 PEM 解码失败")
		}
		// X509 解码, 解析 PKCS1 格式的私钥
		decodedPrivateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, err
		}
		// RSA 解密
		return rsa.DecryptPKCS1v15(rand.Reader, decodedPrivateKey, decodedString)
	}
}

// RSAEncrypt RSA 加密
func RSAEncrypt(origData []byte) (string, error) {
	// 解码 pem 格式的公钥
	block, _ := pem.Decode([]byte(config.Get().RSAPublicKey))
	if block == nil {
		return "", errors.New("RSA 公钥 PEM 解码失败")
	}
	// X509 解码, 解析 PKCS1 格式的公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	// 加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubInterface.(*rsa.PublicKey), origData)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
