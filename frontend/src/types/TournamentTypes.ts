export interface EventApiResponse {
  status: string;
  status_code: number;
  data: EventResponse[];
}

export interface EventUpdate {
  name: string
  start_date: string
  end_date: string
  status: number
}

export interface EventCreate {
  name: string
  start_date: string
  end_date: string
  type_draws: number
}

export interface EventResponse {
  id: string
  name: string
  start_date: string
  end_date: string
  previous_event_id: string | null
  status: number
  type: number
  type_draws: number
  stage: number
  max_stage: number
}

export const EventStatus = {
  Active: 1,
  Closed: 2,
  Registration: 3
} as const