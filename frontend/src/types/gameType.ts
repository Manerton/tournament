export interface GameResponse {
    id: string
    team_a: string
    team_b: string
    event_id: string
    finished: number
}

export interface GameParticipant {
    id: string
    game_id: string
    participant_id: string
    score: number
}

export interface GameAllInformation {
    team_a_id: string
    team_a_name: string
    team_b_id: string
    team_b_name: string
    score_a: number
    score_b: number
    status: number
}