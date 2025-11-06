import axios from "axios";
import { API_URL } from "../config/routes";
import { getCRFCookie } from "../components/helpers/AuthContext";
import { ResponseStatus, type ApiResponse } from "../types/responseTypes";
import type { SubmissionDTOResponse } from "../types/submissionTypes";

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

export async function axiosGetAllSubmissions(id: string) {
  try {
    const res = await api.get<ApiResponse<SubmissionDTOResponse[]>>(`/api/submissions/${id}/all`)
    const json = res.data

    if (json.status == ResponseStatus.Error) {
      throw new Error(json.message || "")
    }

    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "")
  }
}

export async function axiosGetSubmissionByCode(id: string, code: string) {
    try {
    const res = await api.get<ApiResponse<SubmissionDTOResponse>>(`/api/submissions/${id}/one?code=${code}`)
    const json = res.data

    if (json.status == ResponseStatus.Error) {
      throw new Error(json.message || "")
    }

    return json.data
  } catch (error: any) {
    throw new Error(error.response?.data?.message || "")
  }
}