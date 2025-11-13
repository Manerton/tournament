import axios from "axios";
import { API_URL } from "../config/routes";
import { getCRFCookie } from "../components/helpers/AuthContext";

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

export const axiosGetTeamReport = async (tournamentId: string) => {
    const response = await api.get(
        `/api/report/teams-report/${tournamentId}`,
        {
            responseType: "blob", // ⚠️ важно — чтобы axios не пытался прочитать бинарные данные как JSON
        }
    );

    console.log("response", response)
    console.log("API URL", API_URL)


    // создаём ссылку на скачивание файла
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement("a");
    link.href = url;

    // Пытаемся вытащить имя файла из заголовка Content-Disposition
    const disposition = response.headers["content-disposition"];
    let fileName = "report.xlsx";
    if (disposition) {
        const match = disposition.match(/filename="?([^"]+)"?/);
        if (match?.[1]) fileName = match[1];
    }

    link.setAttribute("download", fileName);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
};

export const axiosGetLeaderBoardReport = async (tournamentId: string) => {
    const response = await api.get(
        `/api/report/leaderboard-report/${tournamentId}`,
        {
            responseType: "blob", // ⚠️ важно — чтобы axios не пытался прочитать бинарные данные как JSON
        }
    );

    console.log("response", response)
    console.log("API URL", API_URL)


    // создаём ссылку на скачивание файла
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement("a");
    link.href = url;

    // Пытаемся вытащить имя файла из заголовка Content-Disposition
    const disposition = response.headers["content-disposition"];
    let fileName = "report.xlsx";
    if (disposition) {
        const match = disposition.match(/filename="?([^"]+)"?/);
        if (match?.[1]) fileName = match[1];
    }

    link.setAttribute("download", fileName);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
};