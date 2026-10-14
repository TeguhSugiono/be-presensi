package entity

import "time"

type SchoolLocation struct {
	ID           int     `gorm:"column:id;primaryKey"`
	Name         string  `gorm:"column:name;type:varchar(100);not null"`
	Latitude     float64 `gorm:"column:latitude;type:numeric;not null"`
	Longitude    float64 `gorm:"column:longitude;type:numeric;not null"`
	RadiusMeters int     `gorm:"column:radius_meters;not null;default:100"`
	IsActive     bool    `gorm:"column:is_active;not null;default:true"`

	CreatedBy *int64     `gorm:"column:created_by"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedBy *int64     `gorm:"column:updated_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime:false"`

	IsDeleted bool `gorm:"column:is_deleted;not null;default:false"`
}

func (SchoolLocation) TableName() string {
	return "school_locations"
}
