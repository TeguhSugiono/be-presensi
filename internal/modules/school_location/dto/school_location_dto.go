package dto

type CreateSchoolLocationRequest struct {
	Name         string  `json:"name" binding:"required,max=100"`
	Latitude     float64 `json:"latitude" binding:"required,gte=-90,lte=90"`
	Longitude    float64 `json:"longitude" binding:"required,gte=-180,lte=180"`
	RadiusMeters int     `json:"radius_meters" binding:"required,gt=0"`
	IsActive     bool    `json:"is_active"`
}

type UpdateSchoolLocationRequest struct {
	Name         string  `json:"name" binding:"required,max=100"`
	Latitude     float64 `json:"latitude" binding:"required,gte=-90,lte=90"`
	Longitude    float64 `json:"longitude" binding:"required,gte=-180,lte=180"`
	RadiusMeters int     `json:"radius_meters" binding:"required,gt=0"`
	IsActive     bool    `json:"is_active"`
}

type SchoolLocationResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeters int     `json:"radius_meters"`
	IsActive     bool    `json:"is_active"`
}
