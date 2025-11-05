export interface LeaderboardDTOResponse {
    id: string;
    name: string;
    score: number;
}

export interface LeaderboardPlayersDTOResponse {
    team_name: string;
    name: string;
    score: number;
}

export interface LeaderboardWinDTOResponse {
    id: string;
    name: string;
    scoreWins: number;
    score: number;
}
