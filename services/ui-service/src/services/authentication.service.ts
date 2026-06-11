import type { ILoginResponse, ILoginUser, IRegisterUser } from "../types/user.types";
import { configService } from "./config.service";

// Authentication lives in httpOnly cookies set by the auth-service: login
// sets access_token/refresh_token cookies, every same-origin fetch sends them
// automatically, and JavaScript (including XSS payloads) cannot read them.
// Nothing about the session is persisted in localStorage anymore; the auth
// status is derived by introspecting the cookie against the server.
class AuthenticationService {
  private authStatusListeners: Set<(status: boolean) => void> = new Set();
  private authStatus: boolean | null = null;

  constructor() {
    this.checkInitialAuthStatus();
  }

  private async checkInitialAuthStatus(): Promise<void> {
    const isValid = await this.validateSession();
    this.authStatus = isValid;
    this.notifyAuthStatusChange(isValid);
  }

  subscribeToAuthStatus(callback: (status: boolean) => void): () => void {
    this.authStatusListeners.add(callback);
    if (this.authStatus !== null) {
      callback(this.authStatus);
    }
    return () => this.authStatusListeners.delete(callback);
  }

  private notifyAuthStatusChange(status: boolean): void {
    this.authStatus = status;
    this.authStatusListeners.forEach((callback) => {
      callback(status);
    });
  }

  async login(user: ILoginUser): Promise<ILoginResponse> {
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/sessions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(user),
      });

      if (!response.ok) {
        throw new Error("Login failed");
      }

      const loginResponse: ILoginResponse = await response.json();
      this.notifyAuthStatusChange(true);
      return loginResponse;
    } catch (error) {
      console.error("Login error", error);
      throw error;
    }
  }

  async logout(): Promise<boolean> {
    try {
      // The refresh_token cookie identifies the session; the server clears
      // both auth cookies in the response.
      const response = await fetch(`${configService.authenticationServiceUrl}/sessions/current`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error("Logout failed");
      }

      this.notifyAuthStatusChange(false);
      return true;
    } catch (error) {
      console.error("Logout error", error);
      throw error;
    }
  }

  async register(user: IRegisterUser): Promise<boolean> {
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/users`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(user),
      });

      if (!response.ok) {
        throw new Error("Registration failed");
      }

      await this.login({ email: user.email, password: user.password });

      return true;
    } catch (error) {
      console.error("Registration error", error);
      throw error;
    }
  }

  async isAuthenticated(): Promise<boolean> {
    const isValid = await this.validateSession();
    this.notifyAuthStatusChange(isValid);
    return isValid;
  }

  // validateSession introspects the httpOnly access_token cookie: the body is
  // empty and the server reads the token from the cookie.
  private async validateSession(): Promise<boolean> {
    try {
      const response = await fetch(`${configService.authenticationServiceUrl}/tokens/introspect`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({}),
      });

      if (!response.ok) {
        return false;
      }

      const result = await response.json();
      return result.isValid;
    } catch (error) {
      console.error("Session validation error", error);
      return false;
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
      throw new Error("Failed to fetch paginated users");
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
