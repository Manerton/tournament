package report_handler

import (
	"context"
	"fmt"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ReportServiceInterface interface {
	ExportTeamsToXLSX(ctx context.Context, eventID string) (string, error)
	ExportLeaderboardScoreToXLSX(ctx context.Context, eventId string) (string, error)
}

type ReportHandler struct {
	reportService ReportServiceInterface
}

func New(reportService ReportServiceInterface) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

func (h *ReportHandler) ExportTeamsToXLSX(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	eventId := chi.URLParam(r, "id")

	pathToFile, err := h.reportService.ExportTeamsToXLSX(ctx, eventId)
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

	defer os.Remove(pathToFile)

	fileName := filepath.Base(pathToFile)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("File-Name", fileName)

	http.ServeFile(w, r, pathToFile)
}

func (h *ReportHandler) ExportLeaderBoardToXLSX(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	eventId := chi.URLParam(r, "id")

	pathToFile, err := h.reportService.ExportLeaderboardScoreToXLSX(ctx, eventId)
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

	defer os.Remove(pathToFile)

	fileName := filepath.Base(pathToFile)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("File-Name", fileName)

	http.ServeFile(w, r, pathToFile)
}
