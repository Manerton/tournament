package report_service

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/event/model/event"
	"main/internal/game/model/game"
	"main/internal/game/model/game_participant"
	"main/internal/leaderboard/model/leaderboard"
	"main/internal/location/model/school"
	"main/internal/participant/model/participant"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/report/model/report"
	"main/internal/team/model/team"
	"main/internal/team/model/team_participant"
	"os"
	"sort"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TeamReposittoryInterface interface {
	GetByEventID(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)
	GetAcceptTeamsByEventId(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)
}

type TeamParticipantRepositoryInterface interface {
	GetAllByTeamIDPreload(ctx context.Context, db *gorm.DB, teamID uuid.UUID) ([]team_participant.TeamParticipant, error)
}

type SchoolRepostoryInterface interface {
	GetByIdPrelaod(ctx context.Context, db *gorm.DB, id uuid.UUID) (school.School, error)
}

type LeaderboardRepositoryInterface interface {
	GetByTeamId(ctx context.Context, db *gorm.DB, teamId uuid.UUID) (leaderboard.Leaderboard, error)
	GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]leaderboard.Leaderboard, error)
}

type EventRepositoryInterface interface {
	GetByPreviousEventID(context.Context, *gorm.DB, uuid.UUID) ([]event.Event, error)
}

type GameRepositoryInterface interface {
	GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]game.Game, error)
}

type GameParticipantRepositoryInterface interface {
	GetAllByGameID(ctx context.Context, db *gorm.DB, gameID uuid.UUID) ([]game_participant.GameParticipant, error)
}

type ReportService struct {
	log *slog.Logger
	db  *gorm.DB

	teamRepository            TeamReposittoryInterface
	teamParticipantRepository TeamParticipantRepositoryInterface
	schoolRepository          SchoolRepostoryInterface
	leaderboardRepository     LeaderboardRepositoryInterface
	eventRepository           EventRepositoryInterface
	gameRepository            GameRepositoryInterface
	gameParticipantRepository GameParticipantRepositoryInterface
}

func New(log *slog.Logger, db *gorm.DB, teamRepository TeamReposittoryInterface,
	teamParticipantRepository TeamParticipantRepositoryInterface,
	schoolRepository SchoolRepostoryInterface, leaderboardRepository LeaderboardRepositoryInterface,
	eventRepository EventRepositoryInterface, gameRepository GameRepositoryInterface, gameParticipantRepository GameParticipantRepositoryInterface) *ReportService {
	return &ReportService{
		log:                       log,
		db:                        db,
		teamRepository:            teamRepository,
		teamParticipantRepository: teamParticipantRepository,
		schoolRepository:          schoolRepository,
		leaderboardRepository:     leaderboardRepository,
		eventRepository:           eventRepository,
		gameRepository:            gameRepository,
		gameParticipantRepository: gameParticipantRepository,
	}
}

func (s *ReportService) ExportLeaderboardScoreToXLSX(ctx context.Context, eventId string) (string, error) {
	const op = "service.ReportService.ExportLeaderboardScoreToXLSX"

	log := s.log.With(slog.String("op", op))

	// 1️⃣ Парсим ID
	uid, err := uuid.Parse(eventId)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", eventId), logger.Err(err))
		return "", errs.ErrBadRequest
	}

	// 2️⃣ Получаем агрегированные данные
	participantsReport, err := s.dataAggregationForLeaderboardReport(ctx, uid)
	if err != nil {
		log.Error("failed aggregate data", logger.Err(err))
		return "", errs.ErrInternalError
	}

	// 3️⃣ Заголовки таблицы
	headers := []string{
		"Название команды",
		"ФИО участника",
		"Класс",
		"Школа",
		"Очки участника",
		"Общий счёт команды",
	}

	// 4️⃣ Генерация Excel-файла
	filePath, err := s.fillFileLeaderboardParticipantsToXLSX(headers, participantsReport)
	if err != nil {
		log.Error("failed filling file", logger.Err(err))
		return "", errs.ErrInternalError
	}

	return filePath, nil
}

func (s *ReportService) dataAggregationForLeaderboardReport(ctx context.Context, eventID uuid.UUID) ([]report.LeaderboardPlayesReport, error) {
	// Получаем команды, участвующие в событии
	teams, err := s.teamRepository.GetAcceptTeamsByEventId(ctx, s.db, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams by eventID: %w", err)
	}

	// Получаем все туры, связанные с этим событием
	tours, err := s.eventRepository.GetByPreviousEventID(ctx, s.db, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tours: %w", err)
	}

	// Собираем все игры по турам
	var games []game.Game
	for _, tour := range tours {
		gameInTour, err := s.gameRepository.GetByEventId(ctx, s.db, tour.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get games for tour: %w", err)
		}
		games = append(games, gameInTour...)
	}

	// Подготовка структур для накопления очков
	type participantData struct {
		participant participant.Participant
		score       int
	}
	type teamAggregate struct {
		name         string
		participants map[uuid.UUID]*participantData
	}

	teamStats := make(map[uuid.UUID]*teamAggregate)

	// Инициализация по командам
	for _, t := range teams {
		participants, err := s.teamParticipantRepository.GetAllByTeamIDPreload(ctx, s.db, t.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get participants for team %s: %w", t.Name, err)
		}

		pMap := make(map[uuid.UUID]*participantData)
		for _, tp := range participants {
			pMap[tp.ParticipantID] = &participantData{
				participant: tp.Participant,
				score:       0,
			}
		}

		teamStats[t.ID] = &teamAggregate{
			name:         t.Name,
			participants: pMap,
		}
	}

	// Подсчёт очков
	for _, g := range games {
		gameParticipants, err := s.gameParticipantRepository.GetAllByGameID(ctx, s.db, g.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get game participants: %w", err)
		}

		for _, gp := range gameParticipants {
			for _, team := range teamStats {
				if pdata, exists := team.participants[gp.ParticipantID]; exists {
					pdata.score += gp.Score
					break
				}
			}
		}
	}

	// Формируем отчет
	var reports []report.LeaderboardPlayesReport
	for _, team := range teamStats {
		var totalScore int
		var participantsWithScores []report.ParticipantScore

		for _, pdata := range team.participants {
			totalScore += pdata.score
			participantsWithScores = append(participantsWithScores, report.ParticipantScore{
				FIO:    pdata.participant.FIO,
				Class:  pdata.participant.Class,
				School: pdata.participant.School.Name,
				Score:  pdata.score,
			})
		}

		reports = append(reports, report.LeaderboardPlayesReport{
			TeamName:     team.name,
			Score:        totalScore,
			Participants: participantsWithScores,
		})
	}

	return reports, nil
}

func (s *ReportService) fillFileLeaderboardParticipantsToXLSX(headers []string, data []report.LeaderboardPlayesReport) (string, error) {
	newFile := excelize.NewFile()
	sheet := newFile.GetSheetName(newFile.GetActiveSheetIndex())

	// Заголовки
	for i, header := range headers {
		col := string(rune('A' + i))
		newFile.SetCellValue(sheet, fmt.Sprintf("%s1", col), header)
	}

	// 🔹 Сортируем команды по названию (можно заменить на сортировку по Score, если нужно)
	sort.Slice(data, func(i, j int) bool {
		return data[i].TeamName < data[j].TeamName
	})

	row := 2
	for _, team := range data {
		// Можно также отсортировать участников внутри команды, если нужно, например, по фамилии:
		sort.Slice(team.Participants, func(i, j int) bool {
			return team.Participants[i].FIO < team.Participants[j].FIO
		})

		for _, part := range team.Participants {
			newFile.SetCellValue(sheet, fmt.Sprintf("A%d", row), team.TeamName)
			newFile.SetCellValue(sheet, fmt.Sprintf("B%d", row), part.FIO)
			newFile.SetCellValue(sheet, fmt.Sprintf("C%d", row), part.Class)
			newFile.SetCellValue(sheet, fmt.Sprintf("D%d", row), part.School)
			newFile.SetCellValue(sheet, fmt.Sprintf("E%d", row), part.Score)
			newFile.SetCellValue(sheet, fmt.Sprintf("F%d", row), team.Score)
			row++
		}
	}

	fileName := "export_leaderboard_scores_*.xlsx"
	tmpFile, err := os.CreateTemp("", fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if err := newFile.SaveAs(tmpFile.Name()); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return tmpFile.Name(), nil
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
