package report_service

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/location/model/school"
	"main/internal/participant/model/participant"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/report/model/report"
	"main/internal/team/model/team"
	"main/internal/team/model/team_participant"
	"os"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TeamReposittoryInterface interface {
	GetByEventID(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)
}

type TeamParticipantRepositoryInterface interface {
	GetAllByTeamIDPreload(ctx context.Context, db *gorm.DB, teamID uuid.UUID) ([]team_participant.TeamParticipant, error)
}

type SchoolRepostoryInterface interface {
	GetByIdPrelaod(ctx context.Context, db *gorm.DB, id uuid.UUID) (school.School, error)
}

type ReportService struct {
	log *slog.Logger
	db  *gorm.DB

	teamRepository            TeamReposittoryInterface
	teamParticipantRepository TeamParticipantRepositoryInterface
	schoolRepository          SchoolRepostoryInterface
}

func New(log *slog.Logger, db *gorm.DB, teamRepository TeamReposittoryInterface,
	teamParticipantRepository TeamParticipantRepositoryInterface,
	schoolRepository SchoolRepostoryInterface) *ReportService {
	return &ReportService{
		log:                       log,
		db:                        db,
		teamRepository:            teamRepository,
		teamParticipantRepository: teamParticipantRepository,
		schoolRepository:          schoolRepository,
	}
}

func (s *ReportService) ExportTeamsToXLSX(ctx context.Context, eventId string) (string, error) {
	const op = "service.ReportService.ExportTeamsToXLSX"

	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(eventId)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", eventId), logger.Err(err))
		return "", errs.ErrBadRequest
	}

	teamsReport, err := s.dataAggregationForTeamReport(ctx, uid)
	if err != nil {
		log.Error("failed aggregate data", logger.Err(err))
		return "", errs.ErrInternalError
	}

	headers := []string{
		"Муниципальное образование",
		"Наименование общеобразовательного учреждения",
		"Название команды",
		"ФИО ответсвенного",
		"Номер телефона ответсвенного",
		"Участник 1",
		"Класс",
		"Участник 2",
		"Класс",
		"Участник 3",
		"Класс",
		"Участник 4",
		"Класс",
		"Участник 5",
		"Класс",
	}

	filePath, err := s.fillingFile(headers, teamsReport)
	if err != nil {
		log.Error("failed filling file", logger.Err(err))
		return "", errs.ErrInternalError
	}

	return filePath, nil
}

func (s *ReportService) dataAggregationForTeamReport(ctx context.Context, eventId uuid.UUID) ([]report.TeamReport, error) {

	teamsReport := []report.TeamReport{}

	teams, err := s.teamRepository.GetByEventID(ctx, s.db, eventId)
	if err != nil {
		return nil, fmt.Errorf("failed get teams by eventID: %w", err)
	}

	for _, team := range teams {
		teamReport := report.TeamReport{}

		school, err := s.schoolRepository.GetByIdPrelaod(ctx, s.db, team.SchoolID)
		if err != nil {
			return nil, fmt.Errorf("failed get school: %w", err)
		}

		teamParticipants, err := s.teamParticipantRepository.GetAllByTeamIDPreload(ctx, s.db, team.ID)
		if err != nil {
			return nil, fmt.Errorf("failed get team participants: %w", err)
		}

		teamReport.District = school.District.Name
		teamReport.School = school.Name
		teamReport.TeamName = team.Name
		teamReport.ResponsibleName = team.ResponsibleName
		teamReport.ResponsiblePhone = team.ResponsiblePhone

		participantsTemp := make([]participant.Participant, 0, len(teamParticipants))
		for _, teamParticipant := range teamParticipants {
			participantsTemp = append(participantsTemp, teamParticipant.Participant)
		}

		teamReport.Participants = participantsTemp
		teamsReport = append(teamsReport, teamReport)
	}

	return teamsReport, nil
}

func (s *ReportService) fillingFile(headers []string, data []report.TeamReport) (string, error) {
	newFile := excelize.NewFile()

	actualSheetName := newFile.GetSheetName(newFile.GetActiveSheetIndex())
	for i, header := range headers {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		newFile.SetCellValue(actualSheetName, cell, header)
	}

	offset := 2

	for i, rowData := range data {
		row := i + offset

		newFile.SetCellValue(actualSheetName, fmt.Sprintf("A%d", row), rowData.District)
		newFile.SetCellValue(actualSheetName, fmt.Sprintf("B%d", row), rowData.School)
		newFile.SetCellValue(actualSheetName, fmt.Sprintf("C%d", row), rowData.TeamName)
		newFile.SetCellValue(actualSheetName, fmt.Sprintf("D%d", row), rowData.ResponsibleName)
		newFile.SetCellValue(actualSheetName, fmt.Sprintf("E%d", row), rowData.ResponsiblePhone)
		baseCol := 5 // F — шестая колонка
		for j, participant := range rowData.Participants {
			colFio := colIndexToLetters(baseCol + j*2)
			colClass := colIndexToLetters(baseCol + j*2 + 1)

			newFile.SetCellValue(actualSheetName, fmt.Sprintf("%s%d", colFio, row), participant.FIO)
			newFile.SetCellValue(actualSheetName, fmt.Sprintf("%s%d", colClass, row), participant.Class)
		}
	}

	fileName := "export_teams_*.xlsx"
	tmpFile, err := os.CreateTemp("", fileName)
	if err != nil {
		return "", fmt.Errorf("failed create temp file: %w", err)
	}
	defer tmpFile.Close()

	err = newFile.SaveAs(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed save file: %w", err)
	}

	return tmpFile.Name(), nil
}

func colIndexToLetters(idx int) string {
	result := ""
	for idx >= 0 {
		result = string('A'+(idx%26)) + result
		idx = idx/26 - 1
	}
	return result
}
