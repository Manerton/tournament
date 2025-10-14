// src/api/apiClient.ts
import axios from "axios";
import type { AuthContextType } from "../components/helpers/AuthContext.tsx";

let authCtx: AuthContextType | null = null;

export const setAuthContext = (ctx: AuthContextType) => {
  authCtx = ctx;
};

const API_BASE = import.meta.env.VITE_API_URL || "http://172.16.0.196:8080/api";

export const api = axios.create({
  baseURL: API_BASE,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

// interceptor, который добавляет X-CSRF-Token
api.interceptors.request.use((config) => {
  if (authCtx?.csrfToken && config.url && !config.url.includes("/sessions/login")) {
    config.headers["X-CSRF-Token"] = authCtx.csrfToken;
  }
  return config;
});
