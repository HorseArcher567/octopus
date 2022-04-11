package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"unsafe"
)

func MD5Str(raw []byte) string {
	return fmt.Sprintf("%x", md5.Sum(raw))
}

func Sha1(raw []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(raw))
}

func Sha256(raw []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

func Sha512(raw []byte) string {
	return fmt.Sprintf("%x", sha512.Sum512(raw))
}

func Hmac(h func() hash.Hash, key, raw []byte) string {
	mac := hmac.New(h, key)
	mac.Write(raw)
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func HmacSha1(key, raw []byte) string {
	return Hmac(sha1.New, key, raw)
}

func HmacSha256(key, raw []byte) string {
	return Hmac(sha256.New, key, raw)
}

func HmacSha512(key, raw []byte) string {
	return Hmac(sha512.New, key, raw)
}

/*
 *@func:标准base64编码
 */
func Base64Encode(raw []byte) string {
	return base64.StdEncoding.EncodeToString(raw)
}

func Base64Decode(secret string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(secret)
}

/*
 *@func: 标准base64 url编码
 */
func Base64URLEncode(raw []byte) string {
	return base64.URLEncoding.EncodeToString(raw)
}

func Base64URLDecode(secret string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(secret)
}

/*
 *@note: 返回的[]byte不可修改
 */
func Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func ToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
