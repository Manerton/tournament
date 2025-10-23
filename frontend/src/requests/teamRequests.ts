import axios from "axios";
import { API_URL } from "../config/routes";
import type { DistrictDTOResponse, SchoolDTOResponse } from "../types/locationTypes";
import type { ParticipantResponseDTO } from "../types/participantTypes";
import { ResponseStatus, type ApiResponse } from "../types/responseTypes";
import type { RegisterTeamRequestDTO, TeamResponseDTO } from "../types/teamType";
import { getCRFCookie } from "../components/helpers/AuthContext";



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

export async function GetDistricts() {
    const res = await fetch(`${API_URL}/api/locations/districts`);
    const json: ApiResponse<DistrictDTOResponse[]> = await res.json();
    if (!res.ok) {
        throw new Error(json.message || "Ошибка при загрузке муниципалитетов");
    } else if (json.status == ResponseStatus.Error) {
        throw new Error(json.message);
    }

    return json.data
}

export async function GetSchools(distrcitId: string) {
    const res = await fetch(`${API_URL}/api/locations/districts/${distrcitId}/schools`);
    const json: ApiResponse<SchoolDTOResponse[]> = await res.json();
    if (!res.ok) {
        throw new Error(json.message || "Ошибка при загрузке школ");
    } else if (json.status == ResponseStatus.Error) {
        throw new Error(json.message);
    }
    return json.data
}

export async function RegisterTeam(dto: RegisterTeamRequestDTO) {
    const res = await fetch(`${API_URL}/api/teams/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(dto),
    });

    const json: ApiResponse<any> = await res.json();
    if (!res.ok) {
        throw new Error(json.message || "Ошибка при загрузке школ");
    } else if (json.status == ResponseStatus.Error) {
        throw new Error(json.message);
    }
}

export async function GetTeamsByEvent(Id: string) {
    const res = await fetch(`${API_URL}/api/teams/event/${Id}`);
    const json: ApiResponse<TeamResponseDTO[]> = await res.json();
    if (!res.ok) {
        throw new Error(json.message || "Ошибка при загрузке команд");
    } else if (json.status == ResponseStatus.Error) {
        throw new Error(json.message);
    }
    return json.data
}

export async function GetTeamMembers(id: string) {
    const res = await fetch(`${API_URL}/api/participants/team/${id}`);
    const json: ApiResponse<ParticipantResponseDTO[]> = await res.json();
    if (!res.ok) {
        throw new Error(json.message || "Ошибка при загрузке участников команды")
    } else if (json.status == ResponseStatus.Error) {
        throw new Error(json.message)
    }
    return json.data
}


export async function ChangeTeamStatus(id: string) {
    try {
        const res = await api.post<ApiResponse<any>>(
            `/api/teams/${id}/switch`
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