import axios from "axios";
import type { EventApiResponse} from "../types/TournamentTypes";


// import type {
//   AUTH
// } from "../config/routes.ts"

const API_BASE = import.meta.env.VITE_API_URL || "http://172.16.0.196:8080/api";

const api = axios.create({
  baseURL: API_BASE,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

export const axiosGetTournamentActive = () => {
  return api.get<EventApiResponse>("/events/actual");
};