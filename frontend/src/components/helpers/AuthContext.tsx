import { createContext, useContext, useEffect, useState } from "react";
import type { UserAuth } from "../../types/sessionTypes";
import { useNavigate } from "react-router-dom";
import {
  axiosLoginSession,
  axiosLogoutSession,
  axiosRefreshSession,
  axiosGetSessionById,
} from "../../requests/sessionRequests";

interface AuthContextType {
  user: UserAuth | null;
  initialized: boolean;
  expires?: number;
  login: (login: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<UserAuth | null>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<UserAuth | null>(null);
  const [initialized, setInitialized] = useState(false);
  const [expires, setExpires] = useState<number | undefined>(undefined);
  const navigate = useNavigate();

  /**
   * Авторизация пользователя
   */
  const login = async (login: string, password: string) => {
    try {
      const response = await axiosLoginSession(login, password);
      const userId = response.data.data.user_id;

      // Получаем полную информацию о сессии
      const userSession = await axiosGetSessionById(userId);
      const session = userSession.data.data;

      setUser({
        id: session.user_id,
        email: session.email ?? login,
      });

      // задаём срок жизни токена (30 минут, как на сервере)
      setExpires(Date.now() + 30 * 60 * 1000);
    } catch (err) {
      console.error("Login failed:", err);
      setUser(null);
    }
  };

  /**
   * Обновление session_token и csrf_token
   */
  const refresh = async (): Promise<UserAuth | null> => {
    try {
      console.log("Refreshing session...");
      const response = await axiosRefreshSession();
      console.log("Session refreshed:", response.data.data);

      // задаём новое время истечения (ещё +30 минут)
      setExpires(Date.now() + 30 * 60 * 1000);

      // обновляем пользователя, если он уже есть
      if (user) {
        const userSession = await axiosGetSessionById(user.id);
        const session = userSession.data.data;
        const updatedUser = {
          id: session.user_id,
          email: session.email ?? "",
        };
        setUser(updatedUser);
        return updatedUser;
      }

      return null;
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
      navigate("/login");
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
