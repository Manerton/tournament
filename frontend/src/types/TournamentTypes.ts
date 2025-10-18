export interface EventApiResponse {
  status: string;
  status_code: number;
  data: EventResponse[];
}



export interface EventResponse {
    id:              string
    name:            string
    start_date:       string
    end_date:         string
    previous_event_id: string | null
    status:          number
    type:            number
    type_draws:       number
    stage:           number
}

export const EventStatus = {
  Active: 1,
  Closed: 2
} as const