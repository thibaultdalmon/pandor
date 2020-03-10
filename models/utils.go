package models

import (
	"strings"
	"time"
)

func FormatUID(s string) string {
	uid := "_:" + strings.ToLower(s)
	uid = strings.ReplaceAll(uid, " ", "")
	return uid
}

func FormatTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.000Z", s)
	return t
}
