export interface UserAuth {
  id: string;
  email: string;
}


export interface LoginResponse {
  status: string;
  status_code: number;
  data: {
    user_id: string;
  };
  error?: string;
}

export interface RefreshResponse {
  status: string;
  status_code: number;
  data: {
    message: string;
  };
  error?: string;
}

export interface LogoutResponse {
  status: string;
  status_code: number;
  data: {
    message: string;
  };
  error?: string;
}

export interface UserSession {
  id: string;
  user_id: string;
  token: string;
  created_at: string;
  expires_at: string;
  email?: string;
}

export interface SessionResponse {
  status: string;
  status_code: number;
  data: UserSession;
}
