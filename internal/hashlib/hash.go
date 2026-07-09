package hashlib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"tutproj/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func HashJWT(jwtStr string) string {
	hashBytes := sha256.Sum256([]byte(jwtStr))

	return hex.EncodeToString(hashBytes[:])
}

func GeneratePass(pass string) string {
	hashpass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("error generating password", err)
		return ""
	}
	return string(hashpass)
}

func VerifyCredentials(db *gorm.DB, username, password string) bool {
	var user model.UnamePass
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
}
