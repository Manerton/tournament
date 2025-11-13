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

type LeaderboardReport struct {
	TeamName  string
	ScoreWins int
	Score     int
}

type LeaderboardPlayesReport struct {
	TeamName     string
	Score        int
	Participants []ParticipantScore
}

type ParticipantScore struct {
	FIO    string
	Class  int
	School string
	Score  int
}
