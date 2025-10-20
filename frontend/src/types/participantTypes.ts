export interface ParticipantCreateDTO {
  FIO: string;
  gender: number;
  class: number;
  school_id: string;
}


export interface ParticipantResponseDTO {
  id: string;
  FIO: string;
  gender: number;
  class: number;
}