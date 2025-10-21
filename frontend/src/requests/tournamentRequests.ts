import axios from "axios";
import type { EventApiResponse, EventResponse } from "../types/TournamentTypes";
import { API_URL } from "../config/routes";
import { ResponseStatus, type ApiResponse } from "../types/responseTypes";


// import type {
//   AUTH
// } from "../config/routes.ts"

const API_BASE = API_URL || "http://172.16.0.196:8080";

const api = axios.create({
  baseURL: API_BASE,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

export const axiosGetTournamentActive = () => {
  return api.get<EventApiResponse>("/api/events/actual");
};

export async function axiosGetEventById(id: string) {
  try {
    const res = await api.get<ApiResponse<EventResponse>>(`/api/events/${id}`);
    const json = res.data;

    if (json.status === ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при загрузке события");
    }

    return json.data;
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при загрузке события");
  }
}

export async function axiosGetAllTournaments() {
  try {
    const res = await api.get<ApiResponse<EventResponse[]>>(`/api/events/tournaments`)
    const json = res.data

    if (json.status === ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при загрузке турниров");
    }
    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при загрузке турниров");
  }
} 