export const ResponseStatus = {
    Error: "Error",
    Ok: "Ok"
} as const;


export interface ApiResponse<T> {
  status: string;
  status_code: number;
  data: T;
  message?: string;
}

