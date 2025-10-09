package location_handler

import (
	"context"
	location_dto "main/internal/location/dto"
	"net/http"
)

type DistrictServiceInterface interface {
	GetAll(context.Context) ([]location_dto.DistrictDTO, error)
}

type SchoolServiceInterface interface {
	GetByDistrict(context.Context, string) []location_dto.SchoolDTO
}

type LocationHandler struct {
	districtService DistrictServiceInterface
	schoolService   SchoolServiceInterface
}

func New(districtService DistrictServiceInterface, schoolService SchoolServiceInterface) *LocationHandler {
	return &LocationHandler{
		districtService: districtService,
		schoolService:   schoolService,
	}
}

func (h *LocationHandler) GetAllDistrict(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_ = ctx

}

func (h *LocationHandler) GetSchoolsByDistrictId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_ = ctx

}
