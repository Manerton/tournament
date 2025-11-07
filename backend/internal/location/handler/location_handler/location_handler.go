package location_handler

import (
	"context"
	location_dto "main/internal/location/dto"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type DistrictServiceInterface interface {
	GetAll(context.Context) ([]location_dto.DistrictDTOResponse, error)
}

type SchoolServiceInterface interface {
	GetByDistrictId(ctx context.Context, id string) ([]location_dto.SchoolDTOResponse, error)
	GetByID(ctx context.Context, id string) (location_dto.SchoolDTOResponse, error)
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

	results, err := h.districtService.GetAll(ctx)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       results,
	})

}

func (h *LocationHandler) GetSchoolsByDistrictId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	results, err := h.schoolService.GetByDistrictId(ctx, id)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       results,
	})

}

func (h *LocationHandler) GetById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	results, err := h.schoolService.GetByID(ctx, id)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}
	}
	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       results,
	})
}
