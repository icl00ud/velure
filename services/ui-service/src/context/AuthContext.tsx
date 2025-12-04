import { createContext, useContext, useEffect, useState, useCallback, useMemo, type ReactNode } from "react";
import type { ILoginResponse, ILoginUser, IRegisterUser, Token } from "../types/user.types";
import { configService } from "../services/config.service";

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

const TOKEN_KEY = "token";
const VALIDATION_KEY = "lastValidation";
const VALIDATION_CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<Token | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isInitializing, setIsInitializing] = useState(true);

  const isAuthenticated = useMemo(() => token !== null, [token]);

  // Load token from storage on mount
  useEffect(() => {
    const loadToken = async () => {
      try {
        const storedToken = localStorage.getItem(TOKEN_KEY);
        if (!storedToken) {
          setIsInitializing(false);
          return;
        }

        const parsedToken: Token = JSON.parse(storedToken);

        // Check if recently validated (skip validation if cached)
        const lastValidation = localStorage.getItem(VALIDATION_KEY);
        const now = Date.now();

        if (lastValidation && now - parseInt(lastValidation, 10) < VALIDATION_CACHE_DURATION) {
          setToken(parsedToken);
          setIsInitializing(false);
          return;
        }

        // Validate token with server
        const isValid = await validateTokenWithServer(parsedToken.accessToken);
        if (isValid) {
          setToken(parsedToken);
          localStorage.setItem(VALIDATION_KEY, now.toString());
        } else {
          clearTokenStorage();
        }
      } catch {
        clearTokenStorage();
      } finally {
        setIsInitializing(false);
      }
    };

    loadToken();
  }, []);

  const clearTokenStorage = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(VALIDATION_KEY);
    setToken(null);
  }, []);

  const saveToken = useCallback((newToken: Token) => {
    localStorage.setItem(TOKEN_KEY, JSON.stringify(newToken));
    localStorage.setItem(VALIDATION_KEY, Date.now().toString());
    setToken(newToken);
  }, []);

  const getAccessToken = useCallback(() => {
    return token?.accessToken ?? null;
  }, [token]);

  const login = useCallback(async (user: ILoginUser): Promise<ILoginResponse> => {
    setIsLoading(true);
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(user),
      });

      if (!response.ok) {
        throw new Error("Erro no login");
      }

      const loginResponse: ILoginResponse = await response.json();
      saveToken(loginResponse);
      return loginResponse;
    } finally {
      setIsLoading(false);
    }
  }, [saveToken]);

  const logout = useCallback(async (): Promise<void> => {
    setIsLoading(true);
    try {
      if (token?.refreshToken) {
        await fetch(`${configService.authenticationServiceUrl}/logout/${token.refreshToken}`, {
          method: "DELETE",
        });
      }
    } finally {
      clearTokenStorage();
      setIsLoading(false);
    }
  }, [token, clearTokenStorage]);

  const register = useCallback(async (user: IRegisterUser): Promise<boolean> => {
    setIsLoading(true);
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(user),
      });

      if (!response.ok) {
        throw new Error("Erro no registro");
      }

      // Auto-login after registration
      await login({ email: user.email, password: user.password });
      return true;
    } finally {
      setIsLoading(false);
    }
  }, [login]);

  const value = useMemo(() => ({
    isAuthenticated,
    isLoading,
    isInitializing,
    token,
    login,
    logout,
    register,
    getAccessToken,
  }), [isAuthenticated, isLoading, isInitializing, token, login, logout, register, getAccessToken]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuthContext must be used within an AuthProvider");
  }
  return context;
}

async function validateTokenWithServer(accessToken: string): Promise<boolean> {
  try {
    const response = await fetch(`${configService.authenticationServiceUrl}/validateToken`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token: accessToken }),
    });

    if (!response.ok) return false;

    const result = await response.json();
    return result.isValid;
  } catch {
    return false;
  }
}
