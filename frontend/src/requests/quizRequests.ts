import axios from "axios";
import { API_URL } from "../config/routes";
import { getCRFCookie } from "../components/helpers/AuthContext";
import { ResponseStatus, type ApiResponse } from "../types/responseTypes";
import type {
    BuzzResult,
    GradeResult,
    HeraldLoginResponse,
    QuizPublicState,
    QuizQuestion,
    QuizTeamAdmin,
} from "../types/quizTypes";


const api = axios.create({
    baseURL: API_URL,
    withCredentials: true,
    timeout: 15000,
    headers: {
        "Content-Type": "application/json",
    },
});

api.interceptors.request.use((config) => {
    const csrfToken = getCRFCookie();
    if (csrfToken && config.headers) {
        config.headers["X-CSRF-Token"] = csrfToken;
    }
    return config;
});

function unwrap<T>(json: ApiResponse<T>): T {
    if (json.status === ResponseStatus.Error) {
        throw new Error(json.message || "Ошибка запроса");
    }
    return json.data;
}

// ---------- public ----------

export async function axiosGetQuizState(eventId: string): Promise<QuizPublicState> {
    try {
        const res = await api.get<ApiResponse<QuizPublicState>>(`/api/quiz/state/${eventId}`);
        return unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось загрузить состояние");
    }
}

export async function axiosQuizBuzz(eventId: string, accessCode: string): Promise<BuzzResult> {
    try {
        const res = await api.post<ApiResponse<BuzzResult>>(`/api/quiz/buzz/${eventId}`, {
            access_code: accessCode,
        });
        return unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось зафиксировать нажатие");
    }
}

// ---------- herald ----------

export async function axiosHeraldLogin(eventId: string, accessCode: string): Promise<HeraldLoginResponse> {
    try {
        const res = await api.post<ApiResponse<HeraldLoginResponse>>(`/api/quiz/herald/login/${eventId}`, {
            access_code: accessCode,
        });
        return unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Неверный код команды");
    }
}

export async function axiosHeraldGrade(eventId: string, accessCode: string, grade: string): Promise<GradeResult> {
    try {
        const res = await api.post<ApiResponse<GradeResult>>(`/api/quiz/herald/grade/${eventId}`, {
            access_code: accessCode,
            grade,
        });
        return unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось сохранить оценку");
    }
}

export async function axiosHeraldChooseOrder(eventId: string, accessCode: string, order: string): Promise<void> {
    try {
        const res = await api.post<ApiResponse<void>>(`/api/quiz/herald/order/${eventId}`, {
            access_code: accessCode,
            order,
        });
        unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось выбрать орден");
    }
}

// ---------- admin ----------

export async function axiosQuizSetupTeams(eventId: string): Promise<void> {
    try {
        const res = await api.post<ApiResponse<void>>(`/api/quiz/admin/${eventId}/setup-teams`);
        unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось подготовить команды");
    }
}

export async function axiosQuizGetTeams(eventId: string): Promise<QuizTeamAdmin[]> {
    try {
        const res = await api.get<ApiResponse<QuizTeamAdmin[]>>(`/api/quiz/admin/${eventId}/teams`);
        return unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось загрузить команды");
    }
}

export async function axiosQuizGetQuestions(eventId: string): Promise<QuizQuestion[]> {
    try {
        const res = await api.get<ApiResponse<QuizQuestion[]>>(`/api/quiz/admin/${eventId}/questions`);
        return unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось загрузить вопросы");
    }
}

export async function axiosQuizSaveQuestion(
    eventId: string,
    round: number,
    questionNumber: number,
    text: string,
    answer: string,
): Promise<void> {
    try {
        const res = await api.post<ApiResponse<void>>(`/api/quiz/admin/${eventId}/questions`, {
            round,
            question_number: questionNumber,
            text,
            answer,
        });
        unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось сохранить вопрос");
    }
}

export async function axiosQuizDeleteQuestion(eventId: string, questionId: string): Promise<void> {
    try {
        const res = await api.delete<ApiResponse<void>>(`/api/quiz/admin/${eventId}/questions/${questionId}`);
        unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось удалить вопрос");
    }
}

export async function axiosQuizStartRound(eventId: string, round: number): Promise<void> {
    try {
        const res = await api.post<ApiResponse<void>>(`/api/quiz/admin/${eventId}/round/start`, { round });
        unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось начать раунд");
    }
}

async function adminAction(eventId: string, action: string): Promise<void> {
    try {
        const res = await api.post<ApiResponse<void>>(`/api/quiz/admin/${eventId}/${action}`);
        unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Действие не выполнено");
    }
}

export const axiosQuizStartQuestion = (eventId: string) => adminAction(eventId, "question/start");
export const axiosQuizPauseTimer = (eventId: string) => adminAction(eventId, "question/pause");
export const axiosQuizResumeTimer = (eventId: string) => adminAction(eventId, "question/resume");
export const axiosQuizCloseQuestion = (eventId: string) => adminAction(eventId, "question/close");
export const axiosQuizRevealAnswer = (eventId: string) => adminAction(eventId, "question/reveal");
export const axiosQuizResetQuestion = (eventId: string) => adminAction(eventId, "question/reset");

export async function axiosQuizShowLeaderboard(eventId: string, show: boolean): Promise<void> {
    try {
        const res = await api.post<ApiResponse<void>>(`/api/quiz/admin/${eventId}/leaderboard`, { show });
        unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось переключить таблицу");
    }
}

export async function axiosQuizGrantTurn(eventId: string, teamId: string): Promise<void> {
    try {
        const res = await api.post<ApiResponse<void>>(`/api/quiz/admin/${eventId}/grant`, { team_id: teamId });
        unwrap(res.data);
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Не удалось передать право ответа");
    }
}

// ---------- websocket ----------

export function connectQuizSocket(eventId: string, onState: (state: QuizPublicState) => void): WebSocket {
    const wsUrl = `${API_URL}/api/ws/${eventId}`.replace(/^http/, "ws");
    const socket = new WebSocket(wsUrl);

    socket.onmessage = (event) => {
        try {
            const msg = JSON.parse(event.data);
            if (msg.action === "quiz_state" && msg.payload) {
                onState(msg.payload as QuizPublicState);
            }
        } catch {
            // ignore malformed messages
        }
    };

    return socket;
}
