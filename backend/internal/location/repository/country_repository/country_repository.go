package country_repository

import (
	"main/internal/location/model/country"
	"main/internal/location/repository/location_repository"
)

type CountryRepository struct {
	location_repository.LocationRepositoryBase[country.Country]
}

func New() *CountryRepository {
	return &CountryRepository{}
}
