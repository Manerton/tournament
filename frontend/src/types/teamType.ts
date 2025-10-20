import type { ParticipantCreateDTO } from "./participantTypes";

export interface TeamCreateDTO {
  team_name: string;
  responsible_name: string;
  responsible_number: string;
  event_id: string;
  school_id: string;
}

export interface RegisterTeamRequestDTO {
  team_name: string;
  responsible_name: string;
  responsible_number: string;
  event_id: string;
  school_id: string;
  participants: ParticipantCreateDTO[];
}

export interface TeamResponseDTO {
  id:   string;
  name: string;
  responsible_name: string;
  responsible_phone: string;
  event_id: string;
  school_id: string;
  status:    number; //2 принято, 1 - на рассмотрении
}

export const TeamStatus = {
  Accept: 2,
  Review: 1,
}