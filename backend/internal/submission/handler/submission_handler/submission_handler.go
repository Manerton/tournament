package submission_handler

import (
	"context"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"main/internal/submission/dto/submission_dto"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type SubmissionServiceInterface interface {
	GetByEventID(ctx context.Context, eventId string) ([]submission_dto.SubmissionDTOResponse, error)
	GetByEventIDAndCode(ctx context.Context, eventId string, code string) (submission_dto.SubmissionDTOResponse, error)
}

type SubmissionHandler struct {
	submissionService SubmissionServiceInterface
}

func New(submissionService SubmissionServiceInterface) *SubmissionHandler {
	return &SubmissionHandler{
		submissionService: submissionService,
	}
}

func (h *SubmissionHandler) GetAllSubmissonsByEventId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")

	result, err := h.submissionService.GetByEventID(ctx, id)
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
		Data:       result,
	})
}

func (h *SubmissionHandler) GetSubmissionByEventIdAndCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	code := r.URL.Query().Get("code")
	result, err := h.submissionService.GetByEventIDAndCode(ctx, id, code)
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
		Data:       result,
	})
}
