package models

import "gorm.io/gorm"

type PolicyHistory struct {
	gorm.Model
	UserID              uint   `gorm:"not null"`
	Type                string `gorm:"not null"` // "advise" | "audit"
	InputPromptOrPolicy string `gorm:"type:text;not null"`
	GeneratedPolicy     string `gorm:"type:text"`
	AnalysisReport      string `gorm:"type:text"` // JSON string
}
