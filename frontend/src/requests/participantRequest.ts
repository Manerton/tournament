import axios from "axios";
import { getCRFCookie } from "../components/helpers/AuthContext";
import { API_URL } from "../config/routes";
import { ResponseStatus, type ApiResponse } from "../types/responseTypes";

const api = axios.create({
    baseURL: API_URL,
    withCredentials: true,
    headers: {
        "Content-Type": "application/json",
    },
});

api.interceptors.request.use((config) => {
  const csrfToken =  getCRFCookie();
  if (csrfToken && config.headers) {
    config.headers["X-CSRF-Token"] = csrfToken;
  }
  return config;
});

export async function axiosUpdateParticipantScore(gameId: string, userId:string, delta: number) {
    try {
        const res = await api.post<ApiResponse<any>>(
            `/api/games/vote/${gameId}?participant_id=${userId}&score=${delta}`
        )
        const json = res.data

         if (json.status === ResponseStatus.Error) {
            throw new Error(json.message || "Ошибка при изменении статуса команды");
        }
        return json.data
    } catch (error: any) {
        throw new Error(error.response?.data?.message || "Ошибка при изменении статуса команды");
    }

} 