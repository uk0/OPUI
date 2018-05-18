package passwd

import (
	"crypto/rand"
	"crypto/sha256"
)

// 密码加密功能, 我们用salt和和sha256做高强度的加密
// see https://crackstation.net/hashing-security.htm

type Salt [32]byte
type Passwd Salt

func ZeroSalt() (ret Salt) {
	return
}

func ZeroPasswd() (ret Passwd) {
	return
}

// NewSalt 生成新的随机数
func NewSalt() (ret Salt) {
	if n, err := rand.Read(ret[:]); n != len(ret) || err != nil {
		panic(err)
	}
	return
}

func hash(barePass string, salt Salt) Passwd {
	return Passwd(sha256.Sum256(append(salt[:], []byte(barePass)[:]...)))
}

// 密码加密, 每次都是新的salt
func Encrypt(barePass string) (encrypted Passwd, salt Salt) {
	salt = NewSalt()
	encrypted = hash(barePass, salt)
	return
}

// 验证密码
func Validate(barePass string, encrypted Passwd, salt Salt) bool {
	return hash(barePass, salt) == encrypted
}
