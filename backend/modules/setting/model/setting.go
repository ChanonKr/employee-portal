// Package model holds the GORM structs for app settings; only repository uses them.
package model

import "time"

// Setting is a row of app_settings (see modules/profile/model for the rules).
type Setting struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedBy *int64
	UpdatedAt time.Time
}

func (Setting) TableName() string { return "app_settings" }
