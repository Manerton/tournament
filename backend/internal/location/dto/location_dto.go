package location_dto

type CountryDTOCreate struct {
	Name string `json:"name" binding:"required"`
}

type RegionDTO struct {
	Name      string `json:"name" binding:"required"`
	CountryID string `json:"country_id" binding:"required,uuid"`
}

type DistrictDTO struct {
	Name     string `json:"name" binding:"required"`
	RegionID string `json:"region_id" binding:"required,uuid"`
}

type DistrictDTOResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name" binding:"required"`
	RegionID string `json:"region_id" binding:"required,uuid"`
}

type SchoolDTO struct {
	Name       string `json:"name" binding:"required"`
	DistrictID string `json:"district_id" binding:"required,uuid"`
}

type SchoolDTOResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name" binding:"required"`
	DistrictID string `json:"district_id" binding:"required,uuid"`
}
