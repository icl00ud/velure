import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ILoginUser, IRegisterUser, Token } from "../types/user.types";
import { authenticationService } from "./authentication.service";

// Mock fetch
global.fetch = vi.fn();

// Authentication uses httpOnly cookies set by the server: the browser sends
// them automatically on same-origin requests, so the service never touches
// localStorage and the tests assert the cookie-era contract (empty introspect
// body, body-less logout).
describe("AuthenticationService", () => {
  const mockLoginUser: ILoginUser = {
    email: "test@example.com",
    password: "password123",
  };

  const mockRegisterUser: IRegisterUser = {
    name: "Test User",
    email: "test@example.com",
    password: "password123",
  };

  const mockToken: Token = {
    accessToken: "mock-access-token",
    refreshToken: "mock-refresh-token",
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("login", () => {
    it("should login successfully", async () => {
      const mockResponse = {
        ...mockToken,
        user: { id: 1, email: mockLoginUser.email, name: "Test User" },
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await authenticationService.login(mockLoginUser);

      expect(result).toEqual(mockResponse);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/sessions"),
        expect.objectContaining({ method: "POST" })
      );
      // Session is established by httpOnly cookies; nothing in localStorage.
      expect(localStorage.getItem("token")).toBeNull();
    });

    it("should throw error on failed login", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 401,
      });

      await expect(authenticationService.login(mockLoginUser)).rejects.toThrow("Login failed");
    });

    it("should notify auth status listeners on successful login", async () => {
      const callback = vi.fn();
      authenticationService.subscribeToAuthStatus(callback);

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockToken,
      });

      await authenticationService.login(mockLoginUser);

      expect(callback).toHaveBeenCalledWith(true);
    });

    it("should handle network errors", async () => {
      (global.fetch as any).mockRejectedValueOnce(new Error("Network error"));

      await expect(authenticationService.login(mockLoginUser)).rejects.toThrow();
    });
  });

  describe("logout", () => {
    it("should logout via the refresh_token cookie (no body)", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
      });

      const result = await authenticationService.logout();

      expect(result).toBe(true);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/sessions/current"),
        expect.objectContaining({ method: "DELETE" })
      );
    });

    it("should throw error on failed logout", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 400,
      });

      await expect(authenticationService.logout()).rejects.toThrow("Logout failed");
    });

    it("should notify auth status listeners on logout", async () => {
      const callback = vi.fn();
      authenticationService.subscribeToAuthStatus(callback);

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
      });

      await authenticationService.logout();

      expect(callback).toHaveBeenCalledWith(false);
    });
  });

  describe("register", () => {
    it("should register successfully and auto-login", async () => {
      // First call: register
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ id: 1, email: mockRegisterUser.email, name: mockRegisterUser.name }),
      });
      // Second call: auto-login
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockToken,
      });

      const result = await authenticationService.register(mockRegisterUser);

      expect(result).toBe(true);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/users"),
        expect.objectContaining({
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(mockRegisterUser),
        })
      );
      // Verify auto-login was called
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/sessions"),
        expect.objectContaining({
          method: "POST",
        })
      );
    });

    it("should throw error on failed registration", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 400,
      });

      await expect(authenticationService.register(mockRegisterUser)).rejects.toThrow(
        "Registration failed"
      );
    });
  });

  describe("isAuthenticated", () => {
    it("should introspect the session cookie with an empty body", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ isValid: true }),
      });

      const result = await authenticationService.isAuthenticated();

      expect(result).toBe(true);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/tokens/introspect"),
        expect.objectContaining({ method: "POST", body: JSON.stringify({}) })
      );
    });

    it("should return false when the session is invalid", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ isValid: false }),
      });

      const result = await authenticationService.isAuthenticated();
      expect(result).toBe(false);
    });

    it("should return false when introspection fails", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 400,
      });

      const result = await authenticationService.isAuthenticated();
      expect(result).toBe(false);
    });

    it("should return false on network error", async () => {
      (global.fetch as any).mockRejectedValueOnce(new Error("Network error"));

      const result = await authenticationService.isAuthenticated();
      expect(result).toBe(false);
    });
  });

  describe("getUsersByPage", () => {
    it("should fetch paginated users successfully", async () => {
      const mockUsers = {
        users: [
          { id: 1, name: "User 1", email: "user1@example.com", created_at: "", updated_at: "" },
          { id: 2, name: "User 2", email: "user2@example.com", created_at: "", updated_at: "" },
        ],
        totalCount: 50,
        page: 1,
        pageSize: 10,
        totalPages: 5,
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockUsers,
      });

      const result = await authenticationService.getUsersByPage(1, 10);

      expect(result).toEqual(mockUsers);
      expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining("page=1&pageSize=10"));
    });

    it("should throw error on failed fetch", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
      });

      await expect(authenticationService.getUsersByPage(1, 10)).rejects.toThrow(
        "Failed to fetch paginated users"
      );
    });
  });

  describe("subscribeToAuthStatus", () => {
    it("should call callback with the last known status", async () => {
      // Establish a known status first.
      (global.fetch as any).mockResolvedValueOnce({ ok: true });
      await authenticationService.logout();

      const callback = vi.fn();
      authenticationService.subscribeToAuthStatus(callback);

      expect(callback).toHaveBeenCalledWith(false);
    });

    it("should return unsubscribe function", async () => {
      const callback = vi.fn();
      const unsubscribe = authenticationService.subscribeToAuthStatus(callback);

      expect(typeof unsubscribe).toBe("function");

      callback.mockClear();
      unsubscribe();

      (global.fetch as any).mockResolvedValueOnce({ ok: true });
      await authenticationService.logout();

      expect(callback).not.toHaveBeenCalled();
    });

    it("should notify all subscribers on auth status change", async () => {
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      authenticationService.subscribeToAuthStatus(callback1);
      authenticationService.subscribeToAuthStatus(callback2);

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockToken,
      });

      await authenticationService.login(mockLoginUser);

      expect(callback1).toHaveBeenCalledWith(true);
      expect(callback2).toHaveBeenCalledWith(true);
    });
  });
});
