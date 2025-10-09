import type { ILoginResponse, ILoginUser, IRegisterUser, Token } from "../types/user.types";
import { configService } from "./config.service";

class AuthenticationService {
  private authStatusListeners: Set<(status: boolean) => void> = new Set();
  private validationCacheDuration: number = 5 * 60 * 1000; // 5 minutos

  constructor() {
    this.checkInitialAuthStatus();
  }

  private getLastValidationTime(): number {
    const stored = localStorage.getItem("lastValidation");
    return stored ? parseInt(stored, 10) : 0;
  }

  private setLastValidationTime(time: number): void {
    localStorage.setItem("lastValidation", time.toString());
  }

  private clearLastValidationTime(): void {
    localStorage.removeItem("lastValidation");
  }

  private async checkInitialAuthStatus(): Promise<void> {
    const tokenString = localStorage.getItem("token");
    if (!tokenString) {
      this.notifyAuthStatusChange(false);
      return;
    }

    try {
      const token: Token = JSON.parse(tokenString);
      
      // Se validou recentemente, considera válido sem validar novamente
      const now = Date.now();
      const lastValidation = this.getLastValidationTime();
      
      if (lastValidation > 0 && now - lastValidation < this.validationCacheDuration) {
        this.notifyAuthStatusChange(true);
        return;
      }

      const isValid = await this.validateToken(token);
      this.notifyAuthStatusChange(isValid);
      if (!isValid) {
        localStorage.removeItem("token");
        this.clearLastValidationTime();
      } else {
        this.setLastValidationTime(now);
      }
    } catch (error) {
      console.error("Erro ao verificar status inicial de autenticação:", error);
      localStorage.removeItem("token");
      this.notifyAuthStatusChange(false);
      this.clearLastValidationTime();
    }
  }

  private hasToken(): boolean {
    const tokenString = localStorage.getItem("token");
    return !!tokenString;
  }

  subscribeToAuthStatus(callback: (status: boolean) => void): () => void {
    this.authStatusListeners.add(callback);
    // Envia o status atual imediatamente quando alguém se inscreve
    const tokenString = localStorage.getItem("token");
    callback(!!tokenString);
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
      this.setLastValidationTime(Date.now()); // Marca como validado
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
      this.clearLastValidationTime(); // Reseta o cache de validação
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
      if (!isValid) {
        localStorage.removeItem("token");
      }
      return isValid;
    } catch (error) {
      console.error("Erro ao verificar autenticação:", error);
      localStorage.removeItem("token");
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
        body: JSON.stringify({ token: token.accessToken }),
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

  async getUsersByPage(
    page: number,
    pageSize: number
  ): Promise<{
    users: User[];
    totalCount: number;
    page: number;
    pageSize: number;
    totalPages: number;
  }> {
    const url = `${configService.authenticationServiceUrl}/users?page=${page}&pageSize=${pageSize}`;
    const response = await fetch(url);

    if (!response.ok) {
      throw new Error("Erro ao buscar usuários paginados");
    }

    return response.json();
  }
}

export interface User {
  id: number;
  name: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export const authenticationService = new AuthenticationService();
