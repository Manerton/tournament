import axios from "axios";
import type { EventCreate, EventResponse, EventUpdate } from "../types/TournamentTypes";
import { API_URL } from "../config/routes";
import { ResponseStatus, type ApiResponse } from "../types/responseTypes";
import { getCRFCookie } from "../components/helpers/AuthContext";


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

api.interceptors.request.use((config) => {
  const csrfToken = getCRFCookie();
  if (csrfToken && config.headers) {
    config.headers["X-CSRF-Token"] = csrfToken;
  }
  return config;
});


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


export async function axiosGetEventByStatus(Status: number) {
  try {
    const res = await api.get<ApiResponse<EventResponse[]>>(`/api/events/status/${Status}`);
    const json = res.data;

    if (json.status === ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при загрузке события");
    }

    return json.data;
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при загрузке события");
  }
}


export async function axiosCreateTournament(dto: EventCreate) {
  try {
    const res = await api.post<ApiResponse<any>>(`/api/events/create`, dto)
    const json = res.data;
    if (json.status === ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при создании турнира");
    }

    return json.data;
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при создании турнира");
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

export async function axiosGetTours(id: string) {
  try {
    const res = await api.get<ApiResponse<EventResponse[]>>(`/api/events/${id}/stages`)
    const json = res.data

    if (json.status === ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при загрузке туров");
    }
    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при загрузке туров");
  }
}

export async function axiosFinishTour(id: string) {
  try {
    const res = await api.post<ApiResponse<any>>(`/api/events/${id}/finish`)
    const json = res.data

    if (json.status === ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при завершении тура")
    }
    return json.data

  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при завершении тура")
  }
}

export async function axiosStartTournament(id: string) {
  try {
    const res = await api.post<ApiResponse<any>>(`/api/events/${id}/start`);
    const json = res.data;

    if (json.status === ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при запуске турнира");
    }

    return json.data;
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при запуске турнира");
  }
}

export async function axiosNextTour(id: string) {
  try {
    const res = await api.post<ApiResponse<any>>(`/api/events/${id}/next-tour`)
    const json = res.data

    if (json.status == ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при переходе к следующему туру");
    }

    return json.data;
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при переходе к следующему туру");
  }
}

export async function axiosUpdateTournament(id: string, dto: EventUpdate) {
  try {
    const res = await api.put<ApiResponse<any>>(
      `/api/events/${id}`,
      dto
    )

    const json = res.data;
    if (json.status === ResponseStatus.Error) {
      throw new Error(json.message || "Ошибка при обновлении турнира");
    }

    return json.data;
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при обновлении турнира");
  }
}

export async function axiosDeleteEvent(id: string) {
  try {
    const res = await api.delete<ApiResponse<any>>(
      `/api/events/${id}`
    )

    const json = res.data
    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "Ошибка при удалении события");
  }
}