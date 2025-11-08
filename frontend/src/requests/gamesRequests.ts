import axios from "axios";
import { API_URL } from "../config/routes";
import { getCRFCookie } from "../components/helpers/AuthContext";
import { ResponseStatus, type ApiResponse } from "../types/responseTypes";
import type { GameAllInformation, GameParticipant } from "../types/gameType";


const api = axios.create({
  baseURL: API_URL,
  withCredentials: true,
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

export async function axiosGetGames(id: string) {
  try {
    const res = await api.get<ApiResponse<GameAllInformation[]>>(`/api/games/${id}/all-info`)
    const json = res.data

    if (json.status == ResponseStatus.Error) {
      throw new Error(json.message || "")
    }

    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "")
  }
}

export async function axiosGetGameById(id: string) {
  try {
    const res = await api.get<ApiResponse<GameAllInformation>>(`/api/games/${id}/one`)
    const json = res.data

    if (json.status == ResponseStatus.Error) {
      throw new Error(json.message || "")
    }

    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "")
  }
}

export async function axiosGetGameParticipantsByListId(partisipantIds: string[]) {
  try {
    const res = await api.post<ApiResponse<GameParticipant[]>>(`/api/games/participants/by-list-id`, {
      ids: partisipantIds
    })
    const json = res.data

    if (json.status == ResponseStatus.Error) {
      throw new Error(json.message || "")
    }
    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "")
  }
}

export async function axiosFinishGame(gameId: string) {
  try {
    const res = await api.post<ApiResponse<any>>(`/api/games/${gameId}/finish`)
    const json = res.data
    console.log(json)

    if (json.status == ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при завершении игры")
    }
    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при завершении игры")
  }
}