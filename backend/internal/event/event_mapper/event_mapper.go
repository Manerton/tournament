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
	}
}
