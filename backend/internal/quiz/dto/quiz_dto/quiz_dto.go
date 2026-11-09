package quiz_dto

type RoundConfig struct {
	Round         int    `json:"round"`
	Name          string `json:"name"`
	QuestionCount int    `json:"question_count"`
	TimeSec       int    `json:"time_sec"`
}

type TimerState struct {
	DurationSec int   `json:"duration_sec"`
	RemainingMs int64 `json:"remaining_ms"`
	Running     bool  `json:"running"`
}

type QuestionState struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
	Answer string `json:"answer"`
}

type BuzzEntry struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
	Status   int    `json:"status"`
	IsActive bool   `json:"is_active"`
}

type TeamState struct {
	TeamID             string `json:"team_id"`
	Number             int    `json:"number"`
	Name               string `json:"name"`
	Round1Score        int    `json:"round1_score"`
	Round2Score        int    `json:"round2_score"`
	Round3Score        int    `json:"round3_score"`
	TotalScore         int    `json:"total_score"`
	Round3Order        int    `json:"round3_order"`
	Round3WrongCount   int    `json:"round3_wrong_count"`
	Round3LivesLeft    int    `json:"round3_lives_left"`
	Round3Eliminated   bool   `json:"round3_eliminated"`
}

type PublicState struct {
	EventID          string         `json:"event_id"`
	CurrentRound     int            `json:"current_round"`
	CurrentQuestion  int            `json:"current_question"`
	RoundActive      bool           `json:"round_active"`
	AnswerRevealed   bool           `json:"answer_revealed"`
	LeaderboardShown bool           `json:"leaderboard_shown"`
	Timer            TimerState     `json:"timer"`
	Question         *QuestionState `json:"question"`
	BuzzQueue        []BuzzEntry    `json:"buzz_queue"`
	AttemptsUsed     int            `json:"attempts_used"`
	AttemptsMax      int            `json:"attempts_max"`
	Teams            []TeamState    `json:"teams"`
	Rounds           []RoundConfig  `json:"rounds"`
}

type HeraldLoginResponse struct {
	TeamID string `json:"team_id"`
	Name   string `json:"name"`
	Number int    `json:"number"`
}

type GradeResult struct {
	PointsAwarded int  `json:"points_awarded"`
	Eliminated    bool `json:"eliminated"`
}

type AdminTeam struct {
	ID               string `json:"id"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	AccessCode       string `json:"access_code"`
	Round1Score      int    `json:"round1_score"`
	Round2Score      int    `json:"round2_score"`
	Round3Score      int    `json:"round3_score"`
	Round3Order      int    `json:"round3_order"`
	Round3WrongCount int    `json:"round3_wrong_count"`
	Round3Eliminated bool   `json:"round3_eliminated"`
}

type BuzzResult struct {
	Position int  `json:"position"`
	QueueFull bool `json:"queue_full"`
}
