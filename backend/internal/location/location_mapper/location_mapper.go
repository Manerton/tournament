package location_mapper

import (
	location_dto "main/internal/location/dto"
	"main/internal/location/model/country"
	"main/internal/location/model/district"
	"main/internal/location/model/region"
	"main/internal/location/model/school"

	"github.com/google/uuid"
)

func ToModelCountry(dto location_dto.CountryDTOCreate) country.Country {
	return country.Country{
		Name: dto.Name,
	}
}

func ToModelRegion(dto location_dto.RegionDTO) region.Region {
	countryID, _ := uuid.Parse(dto.CountryID)
	return region.Region{
		Name:      dto.Name,
		CountryID: countryID,
	}
}

func ToModelDistrict(dto location_dto.DistrictDTO) district.District {
	regionID, _ := uuid.Parse(dto.RegionID)
	return district.District{
		Name:     dto.Name,
		RegionID: regionID,
	}
}

func ToDTODistrict(model district.District) location_dto.DistrictDTOResponse {
	return location_dto.DistrictDTOResponse{
		ID:       model.ID.String(),
		Name:     model.Name,
		RegionID: model.RegionID.String(),
	}
}

func ToManyDTODistrict(models []district.District) []location_dto.DistrictDTOResponse {
	results := make([]location_dto.DistrictDTOResponse, 0, len(models))
	for _, model := range models {
		results = append(results, ToDTODistrict(model))
	}
	return results
}

func ToModelSchool(dto location_dto.SchoolDTO) school.School {
	districtID, _ := uuid.Parse(dto.DistrictID)

	return school.School{
		Name:       dto.Name,
		DistrictID: districtID,
	}
}

func ToDTOShcool(model school.School) location_dto.SchoolDTOResponse {
	return location_dto.SchoolDTOResponse{
		ID:         model.ID.String(),
		Name:       model.Name,
		DistrictID: model.DistrictID.String(),
	}
}

func ToManyDTOSchools(models []school.School) []location_dto.SchoolDTOResponse {
	results := make([]location_dto.SchoolDTOResponse, 0, len(models))
	for _, model := range models {
		results = append(results, ToDTOShcool(model))
	}
	return results
}
