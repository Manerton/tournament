import { createContext, useContext, useEffect, useState } from "react";
import type { UserAuth } from "../../types/sessionTypes";
import { useNavigate } from "react-router-dom";
import {
  axiosLoginSession,
  axiosLogoutSession,
  axiosRefreshSession,
  axiosGetSessionById,
} from "../../requests/sessionRequests";
import { setAuthContext } from "../../requests/globalAuth.ts";


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

      // Получаем полную информацию о сессии
      // const userSession = await axiosGetSessionById(userId);
      // const session = userSession.data.data;

      setUser({
        id: userId,
        email: login,
      });

      // задаём срок жизни токена (30 минут, как на сервере)
      setExpires(Date.now() + 30 * 60 * 1000);
    } catch (err) {
      console.error("Login failed:", err);
      setUser(null);
    }
  };


  const hasCookies = (): boolean => {
  return document.cookie.length > 0;
};

  /**
   * Обновление session_token и csrf_token
   */
  const refresh = async (): Promise<UserAuth | null> => {
  
    if (!hasCookies()) {
    console.log("No cookies found, skipping token refresh");
    setInitialized(true);
    return null;
  }
    
    try {
    console.log("Refreshing session...");
    const response = await axiosRefreshSession();
    const refreshedUserId = response.data.data.message;

    console.log("Session refreshed:", refreshedUserId);

    setExpires(Date.now() + 30 * 60 * 1000);

    // Загружаем актуальную сессию по user_id из ответа
    // const userSession = await axiosGetSessionById(refreshedUserId);
    // const session = userSession.data.data;

    const updatedUser = {
      id: refreshedUserId,
      email: "",
    };
    setUser(updatedUser);

    return updatedUser;
  } catch (err) {
    console.error("Session refresh failed:", err);
    setUser(null);
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
    <AuthContext.Provider value={{ user, initialized, expires, login, logout, refresh }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
};
