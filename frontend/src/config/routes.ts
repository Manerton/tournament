export const API_URL = import.meta.env.VITE_API_URL || "/backend";

export const API_CONFIG = {
    AUTH: `${API_URL}/api/users`,
};

export const AUTH = {
    login: `${API_URL}/api/sessions/login`,
    logout: `${API_URL}/api/sessions/logout`,
    refresh: `${API_URL}/api/sessions/refresh`,
    listSessions: `${API_URL}/api/sessions`,
};

export const SCHOOLS = {
    all: `${API_URL}/schools`,
    byId: `${API_URL}/schools/`,
};

export default { API_CONFIG};