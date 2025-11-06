package submission_dto

type SubmissionDTOResponse struct {
	ID      string `json:"id"`
	EventId string `json:"event_id"`
	Code    string `json:"code"`
	Answer  string `json:"answer"`
}
