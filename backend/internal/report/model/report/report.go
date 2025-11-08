package report

import "main/internal/participant/model/participant"

type TeamReport struct {
	District         string
	School           string
	TeamName         string
	ResponsibleName  string
	ResponsiblePhone string
	Participants     []participant.Participant
}
