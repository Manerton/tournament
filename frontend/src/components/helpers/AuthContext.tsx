import { createContext, useContext, useEffect, useState } from "react";
import type { UserAuth } from "../../types/sessionTypes";
//import { useNavigate } from "react-router-dom";
import {
  axiosLoginSession,
  axiosLogoutSession,
  axiosRefreshSession,
  //axiosGetSessionById,
} from "../../requests/sessionRequests";
import { setAuthContext } from "../../requests/globalAuth.ts";

export function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()!.split(";").shift() || null;
  return null;
}

export interface AuthContextType {
  user: UserAuth | null;            // Текущий пользователь
  initialized: boolean;             // Флаг, что контекст инициализирован
  expires?: number;                 // Время истечения сессии (timestamp)
  csrfToken?: string | null;        // Текущий CSRF-токен
  login: (login: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<UserAuth | null>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<UserAuth | null>(null);
  const [csrfToken, setCsrfToken] = useState<string | null>(null);
  const [initialized, setInitialized] = useState(false);
  const [expires, setExpires] = useState<number | undefined>(undefined);

  useEffect(() => {
    // Подключаем контекст к axios
    setAuthContext({
      user,
      initialized,
      expires,
      csrfToken,
      login,
      logout,
      refresh,
    });
  }, [user, csrfToken, initialized, expires]);

  const login = async (login: string, password: string) => {
    try {
      const response = await axiosLoginSession(login, password);
      const userId = response.data.data.user_id;

      setUser({ id: userId, email: login });
      setExpires(Date.now() + 30 * 60 * 1000);

      // Читаем CSRF из cookie после логина
      const token = getCookie("csrf_token");
      if (token) setCsrfToken(token);

    } catch (err) {
      console.error("Login failed:", err);
      setUser(null);
      setCsrfToken(null);
    }
  };

  const hasCookies = () => !!document.cookie;

  /**
   * Обновление session_token и csrf_token
   */
   const refresh = async (): Promise<UserAuth | null> => {
    if (!hasCookies()) {
      console.log("No cookies found, skipping refresh");
      setInitialized(true);
      return null;
    }

    try {
      const response = await axiosRefreshSession();
      const refreshedUserId = response.data.data.message;

      setExpires(Date.now() + 30 * 60 * 1000);

      // Обновляем CSRF токен после refresh
      const token = getCookie("csrf_token");
      if (token) setCsrfToken(token);

      const updatedUser = { id: refreshedUserId, email: "" };
      setUser(updatedUser);
      return updatedUser;

    } catch (err) {
      console.error("Session refresh failed:", err);
      setUser(null);
      setCsrfToken(null);
      return null;
    } finally {
      setInitialized(true);
    }
  };

  /**
   * Выход из системы
   */
  const logout = async () => {
    try {
      if (user) {
        await axiosLogoutSession(user.id);
      }
    } catch (err) {
      console.error("Logout failed:", err);
    } finally {
      setUser(null);
    }
  };

  /**
   * Плановое обновление токена за минуту до истечения
   */
  useEffect(() => {
    if (!expires) return;

    const timeout = expires - Date.now() - 60_000; // за 1 минуту до истечения

    if (timeout > 0) {
      const id = setTimeout(() => {
        console.log("Auto-refreshing session...");
        refresh();
      }, timeout);

      return () => clearTimeout(id);
    }
  }, [expires]);

  /**
   * При монтировании проверяем текущую сессию
   */
  useEffect(() => {
    refresh();
  }, []);

  return (
    <AuthContext.Provider value={{ user, initialized, expires, csrfToken, login, logout, refresh }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
};
