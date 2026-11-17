package setting

import "time"

// GORM struct for app_settings (see profile/model.go for the rules).
type settingModel struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedBy *int64
	UpdatedAt time.Time
}

func (settingModel) TableName() string { return "app_settings" }
