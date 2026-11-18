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

export const axiosGetTables = async (): Promise<string[]> => {
  const response = await api.get("/api/db-admin/tables");
  return response.data;
};

export const axiosExportTables = async (table: string, format: string) => {
  const response = await api.get(`/api/db-admin/export?table=${table}&format=${format}`, {
    responseType: "blob", // ⚠️ важно — чтобы axios не пытался прочитать бинарные данные как JSON
  });

  // создаём ссылку на скачивание файла
  const url = window.URL.createObjectURL(new Blob([response.data]));
  const link = document.createElement("a");
  link.href = url;

  // Пытаемся вытащить имя файла из заголовка Content-Disposition
  const disposition = response.headers["content-disposition"];
  let fileName = "db-export.json";
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

export interface DbExportFile {
  exported_at?: string;
  tables: Record<string, Record<string, unknown>[]>;
}

export const axiosImportTables = async (data: DbExportFile) => {
  const response = await api.post("/api/db-admin/import", data);
  return response.data;
};
