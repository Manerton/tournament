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