package event_mapper

import (
	"main/internal/event/dto/event_dto"
	"main/internal/event/model/event"
	"time"

	"github.com/google/uuid"
)

func ToModelEvent(dto event_dto.EventDTOCreate) event.Event {
	startDate, _ := time.Parse("2006-01-02", dto.StartDate)
	endDate, _ := time.Parse("2006-01-02", dto.EndDate)

	var prev *uuid.UUID
	if dto.PreviousEventID != nil {
		if parsed, err := uuid.Parse(*dto.PreviousEventID); err == nil {
			prev = &parsed
		}
	}

	return event.Event{
		Name:            dto.Name,
		StartDate:       startDate,
		EndDate:         endDate,
		TypeDraws:       dto.TypeDraws,
		PreviousEventID: prev,
		Status:          event.ActiveStatus,
	}
}

func ToDTOEvent(model event.Event) event_dto.EventDTOResponse {
	var prevId string
	if model.PreviousEventID != nil {
		prevId = model.PreviousEventID.String()
	}

	return event_dto.EventDTOResponse{
		ID:              model.ID.String(),
		Name:            model.Name,
		StartDate:       model.StartDate.String(),
		EndDate:         model.EndDate.String(),
		PreviousEventID: prevId,
		Status:          model.Status,
		Type:            model.Type,
		TypeDraws:       model.TypeDraws,
		Stage:           model.Stage,
	}
}

func ToDTOEvents(models []event.Event) []event_dto.EventDTOResponse {
	dtos := make([]event_dto.EventDTOResponse, 0, len(models))
	for i, model := range models {
		dtos[i] = ToDTOEvent(model)
	}
	return dtos
}
