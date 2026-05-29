package paste

import (
	"crypto/md5"
	"encoding/base64"
	"time"
)

type Visibility int

const (
	VisibilityPublic Visibility = iota + 1
	VisibilityUnlisted
)

type Model struct {
	Id         string
	Content    string
	ExpireAt   time.Time
	Visibility Visibility
}

// Convert to MD5 and later to 22 char URL safe output
func GetID(content string) string {
	hash := md5.Sum([]byte(content))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
