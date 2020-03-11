package models

import (
	"time"
)

// FormatUID builds an unused UID for Dgraph
func FormatUID(name string) string {
	uid := "_:" + name
	return uid
}

// FormatTime converts a string to a time.Time
func FormatTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.000Z", s)
	return t
}
