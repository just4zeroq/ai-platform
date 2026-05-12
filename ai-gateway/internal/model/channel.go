package model

import "gorm.io/gorm"

type Channel struct {
	Id           int             `json:"id" gorm:"primaryKey"`
	Type         int             `json:"type" gorm:"type:int;default:0"`
	Key          string          `json:"key" gorm:"type:text"`
	Name         string          `json:"name" gorm:"type:varchar(128)"`
	Models       string          `json:"models" gorm:"type:text"`
	Group        string          `json:"group" gorm:"type:varchar(64);default:'default'"`
	Status       int             `json:"status" gorm:"type:int;default:1"`
	Priority     *int64          `json:"priority" gorm:"bigint;default:0"`
	Weight       uint            `json:"weight" gorm:"default:0"`
	ModelMapping string          `json:"model_mapping" gorm:"type:text"`
	BaseUrl      string          `json:"base_url" gorm:"type:varchar(255)"`
	CreatedAt    int64           `json:"created_at" gorm:"bigint"`
	UpdatedAt    int64           `json:"updated_at" gorm:"bigint"`
	DeletedAt    gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}
