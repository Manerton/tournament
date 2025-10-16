import { API_URL } from "../config/routes";
import type { DistrictDTOResponse, SchoolDTOResponse } from "../types/locationTypes";
import { ResponseStatus, type ApiResponse } from "../types/responseTypes";
import type { RegisterTeamRequestDTO } from "../types/teamType";


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