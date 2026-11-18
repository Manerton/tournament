// src/requests/globalAuth.ts
import axios from "axios";
import type { AxiosError, InternalAxiosRequestConfig } from "axios";
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

function readCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()!.split(";").shift() || null;
  return null;
}

type RetryableConfig = InternalAxiosRequestConfig & { _authRetried?: boolean };

/**
 * Response interceptor: on 401/403 awaits the single-flight session refresh
 * and retries the original request once. Safe to attach to any axios instance.
 */
export function attachRefreshRetry(instance: typeof api) {
  instance.interceptors.response.use(undefined, async (error: AxiosError) => {
    const config = error.config as RetryableConfig | undefined;
    const status = error.response?.status;

    if (!config || config._authRetried) return Promise.reject(error);
    if (status !== 401 && status !== 403) return Promise.reject(error);
    if ((config.url ?? "").includes("/sessions/")) return Promise.reject(error);
    if (!readCookie("session_token")) return Promise.reject(error);
    if (!authCtx?.refresh) return Promise.reject(error);

    config._authRetried = true;
    try {
      const user = await authCtx.refresh();
      if (!user) return Promise.reject(error);
      return instance(config);
    } catch {
      return Promise.reject(error);
    }
  });
}

// interceptor, который добавляет X-CSRF-Token из cookie на момент запроса
api.interceptors.request.use((config) => {
  const csrfToken = readCookie("csrf_token");
  if (csrfToken && config.url && !config.url.includes("/sessions/")) {
    config.headers["X-CSRF-Token"] = csrfToken;
  }
  return config;
});

attachRefreshRetry(api);
