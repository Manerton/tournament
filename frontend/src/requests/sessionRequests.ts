import axios from "axios";
import type {
  LoginResponse,
  RefreshResponse,
  LogoutResponse,
  SessionResponse,
} from "../types/sessionTypes";

// import type {
//   AUTH
// } from "../config/routes.ts"

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

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
  return api.post<LoginResponse>("/sessions/login", { login, password });
};

/**
 * POST /sessions/refresh
 */
export const axiosRefreshSession = () => {
  return api.post<RefreshResponse>("/sessions/refresh");
};

/**
 * POST /sessions/logout/{id}
 */
export const axiosLogoutSession = (id: string) => {
  return api.post<LogoutResponse>(`/sessions/logout/${id}`);
};

/**
 * GET /sessions/{id}
 */
export const axiosGetSessionById = (id: string) => {
  return api.get<SessionResponse>(`/sessions/${id}`);
};
