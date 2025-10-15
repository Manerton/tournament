import axios from "axios";
import type {
  LoginResponse,
  RefreshResponse,
  LogoutResponse,
  SessionResponse,
} from "../types/sessionTypes";
import { API_URL } from "../config/routes";

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

/**
 * POST /sessions/login
 */
export const axiosLoginSession = (login: string, password: string) => {
  return api.post<LoginResponse>("/api/sessions/login", { login, password });
};

/**
 * POST /sessions/refresh
 */
export const axiosRefreshSession = () => {
  return api.post<RefreshResponse>("/api/sessions/refresh");
};

/**
 * POST /sessions/logout/{id}
 */
export const axiosLogoutSession = (id: string) => {
  return api.post<LogoutResponse>(`/api/sessions/logout/${id}`);
};

/**
 * GET /sessions/{id}
 */
export const axiosGetSessionById = (id: string) => {
  return api.get<SessionResponse>(`/api/sessions/${id}`);
};
