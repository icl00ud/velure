import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { configService } from "../services/config.service";
import type { ILoginResponse, ILoginUser, IRegisterUser, Token } from "../types/user.types";

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  isInitializing: boolean;
  token: Token | null;
  login: (user: ILoginUser) => Promise<ILoginResponse>;
  logout: () => Promise<void>;
  register: (user: IRegisterUser) => Promise<boolean>;
  getAccessToken: () => string | null;
}

const AuthContext = createContext<AuthContextType | null>(null);

// Authentication lives in httpOnly cookies set by the auth-service: login
// sets access_token/refresh_token cookies, every same-origin fetch sends them
// automatically, and JavaScript (including XSS payloads) cannot read them.
// The token kept in React state is in-memory only and never persisted.
export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<Token | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isInitializing, setIsInitializing] = useState(true);

  useEffect(() => {
    const restoreSession = async () => {
      try {
        // Introspect with an empty body: the server validates the
        // access_token cookie sent along with the request.
        setIsAuthenticated(await validateSessionWithServer());
      } catch {
        setIsAuthenticated(false);
      } finally {
        setIsInitializing(false);
      }
    };

    restoreSession();
  }, []);

  const getAccessToken = useCallback(() => {
    return token?.accessToken ?? null;
  }, [token]);

  const login = useCallback(async (user: ILoginUser): Promise<ILoginResponse> => {
    setIsLoading(true);
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/sessions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(user),
      });

      if (!response.ok) {
        throw new Error("Erro no login");
      }

      const loginResponse: ILoginResponse = await response.json();
      setToken(loginResponse);
      setIsAuthenticated(true);
      return loginResponse;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = useCallback(async (): Promise<void> => {
    setIsLoading(true);
    try {
      // The refresh_token cookie identifies the session; the server clears
      // both auth cookies in the response.
      await fetch(`${configService.authenticationServiceUrl}/sessions/current`, {
        method: "DELETE",
      });
    } finally {
      setToken(null);
      setIsAuthenticated(false);
      setIsLoading(false);
    }
  }, []);

  const register = useCallback(
    async (user: IRegisterUser): Promise<boolean> => {
      setIsLoading(true);
      try {
        const response = await fetch(`${configService.authenticationServiceUrl}/users`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(user),
        });

        if (!response.ok) {
          throw new Error("Erro no registro");
        }

        await login({ email: user.email, password: user.password });
        return true;
      } finally {
        setIsLoading(false);
      }
    },
    [login]
  );

  const value = useMemo(
    () => ({
      isAuthenticated,
      isLoading,
      isInitializing,
      token,
      login,
      logout,
      register,
      getAccessToken,
    }),
    [isAuthenticated, isLoading, isInitializing, token, login, logout, register, getAccessToken]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuthContext must be used within an AuthProvider");
  }
  return context;
}

async function validateSessionWithServer(): Promise<boolean> {
  try {
    const response = await fetch(`${configService.authenticationServiceUrl}/tokens/introspect`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });

    if (!response.ok) return false;

    const result = await response.json();
    return result.isValid;
  } catch {
    return false;
  }
}
