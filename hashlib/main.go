package hashlib

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashJWT(jwtStr string) string {
	// Convert string to bytes and compute SHA-256 hash
	hashBytes := sha256.Sum256([]byte(jwtStr))

	// Encode the byte array to a readable hexadecimal string
	return hex.EncodeToString(hashBytes[:])
}
