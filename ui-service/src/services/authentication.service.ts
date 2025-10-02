import type { ILoginResponse, ILoginUser, IRegisterUser, Token } from "../types/user.types";
import { configService } from "./config.service";

class AuthenticationService {
  private authStatusListeners: Set<(status: boolean) => void> = new Set();

  constructor() {
    this.checkInitialAuthStatus();
  }

  private async checkInitialAuthStatus(): Promise<void> {
    const isAuth = await this.isAuthenticated();
    this.notifyAuthStatusChange(isAuth);
  }

  private hasToken(): boolean {
    const tokenString = localStorage.getItem("token");
    return !!tokenString;
  }

  subscribeToAuthStatus(callback: (status: boolean) => void): () => void {
    this.authStatusListeners.add(callback);
    return () => this.authStatusListeners.delete(callback);
  }

  private notifyAuthStatusChange(status: boolean): void {
    this.authStatusListeners.forEach((callback) => callback(status));
  }

  async login(user: ILoginUser): Promise<ILoginResponse> {
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(user),
      });

      if (!response.ok) {
        throw new Error("Erro no login");
      }

      const loginResponse: ILoginResponse = await response.json();
      localStorage.setItem("token", JSON.stringify(loginResponse));
      this.notifyAuthStatusChange(true);
      console.log("Login realizado com sucesso.");
      return loginResponse;
    } catch (error) {
      console.error("Erro no login", error);
      throw error;
    }
  }

  async logout(refreshToken: string): Promise<boolean> {
    try {
      const response = await fetch(
        `${configService.authenticationServiceUrl}/logout/${refreshToken}`,
        {
          method: "DELETE",
        }
      );

      if (!response.ok) {
        throw new Error("Erro no logout");
      }

      localStorage.removeItem("token");
      this.notifyAuthStatusChange(false);
      console.log("Logout realizado com sucesso.");
      return true;
    } catch (error) {
      console.error("Erro no logout", error);
      throw error;
    }
  }

  async register(user: IRegisterUser): Promise<boolean> {
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/register`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(user),
      });

      if (!response.ok) {
        throw new Error("Erro no registro");
      }

      console.log("Registro realizado com sucesso.");
      return true;
    } catch (error) {
      console.error("Erro no registro", error);
      throw error;
    }
  }

  async isAuthenticated(): Promise<boolean> {
    const tokenString = localStorage.getItem("token");
    if (!tokenString) {
      return false;
    }

    try {
      const token: Token = JSON.parse(tokenString);
      const isValid = await this.validateToken(token);
      this.notifyAuthStatusChange(isValid);
      return isValid;
    } catch (error) {
      this.notifyAuthStatusChange(false);
      return false;
    }
  }

  private async validateToken(token: Token): Promise<boolean> {
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/validateToken`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ token }),
      });

      if (!response.ok) {
        return false;
      }

      const result = await response.json();
      return result.isValid;
    } catch (error) {
      console.error("Erro na validação do token", error);
      return false;
    }
  }

  getStoredToken(): Token | null {
    const tokenString = localStorage.getItem("token");
    if (!tokenString) return null;

    try {
      return JSON.parse(tokenString);
    } catch (error) {
      return null;
    }
  }
}

export const authenticationService = new AuthenticationService();
