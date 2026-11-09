export interface QuizRoundConfig {
    round: number;
    name: string;
    question_count: number;
    time_sec: number;
}

export interface QuizTimerState {
    duration_sec: number;
    remaining_ms: number;
    running: boolean;
}

export interface QuizQuestionState {
    number: number;
    text: string;
    answer: string;
}

export interface QuizBuzzEntry {
    team_id: string;
    team_name: string;
    status: number; // 1 pending, 2 correct, 3 wrong, 4 false start
    is_active: boolean;
}

export interface QuizTeamState {
    team_id: string;
    number: number;
    name: string;
    round1_score: number;
    round2_score: number;
    round3_score: number;
    total_score: number;
    round3_order: number; // 0 none, 1 red, 2 yellow, 3 green
    round3_wrong_count: number;
    round3_lives_left: number;
    round3_eliminated: boolean;
}

export interface QuizPublicState {
    event_id: string;
    current_round: number;
    current_question: number;
    round_active: boolean;
    answer_revealed: boolean;
    leaderboard_shown: boolean;
    timer: QuizTimerState;
    question: QuizQuestionState | null;
    buzz_queue: QuizBuzzEntry[];
    attempts_used: number;
    attempts_max: number;
    teams: QuizTeamState[];
    rounds: QuizRoundConfig[];
}

export interface QuizQuestion {
    id: string;
    event_id: string;
    round: number;
    question_number: number;
    text: string;
    answer: string;
}

export interface HeraldLoginResponse {
    team_id: string;
    name: string;
    number: number;
}

export interface GradeResult {
    points_awarded: number;
    eliminated: boolean;
}

export interface BuzzResult {
    position: number;
    queue_full: boolean;
}

export interface QuizTeamAdmin {
    id: string;
    team_id: string;
    name: string;
    access_code: string;
    round1_score: number;
    round2_score: number;
    round3_score: number;
    round3_order: number;
    round3_wrong_count: number;
    round3_eliminated: boolean;
}

export const BuzzStatus = {
    Pending: 1,
    Correct: 2,
    Wrong: 3,
    FalseStart: 4,
} as const;

export const Round3OrderNames: Record<number, string> = {
    1: "Красный",
    2: "Жёлтый",
    3: "Зелёный",
};
