package auth

import "golang.org/x/crypto/bcrypt"

// Hash 加密密码（兼容 Java BCrypt 输出格式）。
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// Verify 校验密码。
func Verify(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
