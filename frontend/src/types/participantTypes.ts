export interface ParticipantCreateDTO {
  FIO: string;
  gender: number;
  class: number;
  school_id: string;
  initial_score: number;
}


export interface ParticipantResponseDTO {
  id: string;
  FIO: string;
  gender: number;
  class: number;
  initial_score: number;
}