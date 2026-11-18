package dbadmin_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

const maxImportBodySize = 100 << 20 // 100 MB

type DbAdminServiceInterface interface {
	ListTables(ctx context.Context) []string
	Export(ctx context.Context, table string) (map[string][]map[string]any, error)
	ExportSQL(ctx context.Context, table string) (string, error)
	Import(ctx context.Context, tables map[string][]map[string]any) error
}

type DbAdminHandler struct {
	dbAdminService DbAdminServiceInterface
}

func New(dbAdminService DbAdminServiceInterface) *DbAdminHandler {
	return &DbAdminHandler{
		dbAdminService: dbAdminService,
	}
}

type exportResponse struct {
	ExportedAt time.Time                   `json:"exported_at"`
	Tables     map[string][]map[string]any `json:"tables"`
}

type importRequest struct {
	Tables map[string][]map[string]any `json:"tables"`
}

func (h *DbAdminHandler) ListTables(w http.ResponseWriter, r *http.Request) {
	tables := h.dbAdminService.ListTables(r.Context())
	render.JSON(w, r, tables)
}

func (h *DbAdminHandler) Export(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	table := r.URL.Query().Get("table")
	if table == "" {
		table = "all"
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	stamp := time.Now().UTC().Format("20060102-150405")

	if format == "sql" {
		h.exportSQL(w, r, ctx, table, stamp)
		return
	}

	if format != "json" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap(fmt.Sprintf("unknown format %q, expected json or sql", format))))
		return
	}

	data, err := h.dbAdminService.Export(ctx, table)
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

	payload, err := json.Marshal(exportResponse{
		ExportedAt: time.Now().UTC(),
		Tables:     data,
	})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	fileName := fmt.Sprintf("db-export-%s.json", stamp)
	if table != "all" {
		fileName = fmt.Sprintf("%s-%s.json", table, stamp)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("File-Name", fileName)
	w.Write(payload)
}

func (h *DbAdminHandler) exportSQL(w http.ResponseWriter, r *http.Request, ctx context.Context, table string, stamp string) {
	script, err := h.dbAdminService.ExportSQL(ctx, table)
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

	fileName := fmt.Sprintf("db-export-%s.sql", stamp)
	if table != "all" {
		fileName = fmt.Sprintf("%s-%s.sql", table, stamp)
	}

	w.Header().Set("Content-Type", "application/sql")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("File-Name", fileName)
	w.Write([]byte(script))
}

func (h *DbAdminHandler) Import(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, maxImportBodySize)

	var req importRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap("invalid JSON body")))
		return
	}

	if err := h.dbAdminService.Import(ctx, req.Tables); err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.JSON(w, r, response.SuccessResponse("import completed"))
}
