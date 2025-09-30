package country_repository

import (
	"main/internal/common/repository"
	"main/internal/location/model/country"
)

type CountryRepository struct {
	repository.RepositoryBase[country.Country]
}

func New() *CountryRepository {
	return &CountryRepository{}
}
