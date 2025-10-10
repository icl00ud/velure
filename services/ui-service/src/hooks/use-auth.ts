import { useEffect, useState } from "react";
import { authenticationService } from "../services/authentication.service";
import type { ILoginResponse, ILoginUser, IRegisterUser } from "../types/user.types";

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isInitializing, setIsInitializing] = useState<boolean>(true);

  useEffect(() => {
    const unsubscribe = authenticationService.subscribeToAuthStatus((status) => {
      setIsAuthenticated(status);
      setIsInitializing(false);
    });

    return unsubscribe;
  }, []);

  const login = async (user: ILoginUser): Promise<ILoginResponse> => {
    setIsLoading(true);
    try {
      const result = await authenticationService.login(user);
      return result;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    setIsLoading(true);
    try {
      const token = authenticationService.getStoredToken();
      if (token) {
        await authenticationService.logout(token.refreshToken);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (user: IRegisterUser): Promise<boolean> => {
    setIsLoading(true);
    try {
      return await authenticationService.register(user);
    } finally {
      setIsLoading(false);
    }
  };

  return {
    isAuthenticated,
    isLoading,
    isInitializing,
    login,
    logout,
    register,
  };
}
