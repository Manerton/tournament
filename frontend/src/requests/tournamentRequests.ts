import axios from "axios";
import type { EventApiResponse} from "../types/TournamentTypes";
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

export const axiosGetTournamentActive = () => {
  return api.get<EventApiResponse>("/api/events/actual");
};