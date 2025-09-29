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

func ToModelSchool(dto location_dto.SchoolDTO) school.School {
	districtID, _ := uuid.Parse(dto.DistrictID)

	return school.School{
		Name:       dto.Name,
		DistrictID: districtID,
	}
}
