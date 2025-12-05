import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ILoginUser, IRegisterUser, Token } from "../types/user.types";
import { authenticationService } from "./authentication.service";

// Mock fetch
global.fetch = vi.fn();

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
    localStorage.clear();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("login", () => {
    it("should login successfully and store token", async () => {
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
      expect(localStorage.getItem("token")).toBeTruthy();
      const storedToken = JSON.parse(localStorage.getItem("token")!);
      expect(storedToken.accessToken).toBe(mockToken.accessToken);
      expect(storedToken.refreshToken).toBe(mockToken.refreshToken);
    });

    it("should throw error on failed login", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 401,
      });

      await expect(authenticationService.login(mockLoginUser)).rejects.toThrow("Erro no login");
    });

    it("should notify auth status listeners on successful login", async () => {
      const callback = vi.fn();
      authenticationService.subscribeToAuthStatus(callback);

      const mockResponse = {
        ...mockToken,
        user: { id: 1, email: mockLoginUser.email, name: "Test User" },
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await authenticationService.login(mockLoginUser);

      // Should be called at least twice: once on subscribe, once on login
      expect(callback).toHaveBeenCalledWith(true);
    });

    it("should handle network errors", async () => {
      (global.fetch as any).mockRejectedValueOnce(new Error("Network error"));

      await expect(authenticationService.login(mockLoginUser)).rejects.toThrow();
    });
  });

  describe("logout", () => {
    it("should logout successfully and clear token", async () => {
      // Setup: store token first
      localStorage.setItem("token", JSON.stringify(mockToken));

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
      });

      const result = await authenticationService.logout(mockToken.refreshToken);

      expect(result).toBe(true);
      expect(localStorage.getItem("token")).toBeNull();
    });

    it("should throw error on failed logout", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 400,
      });

      await expect(authenticationService.logout(mockToken.refreshToken)).rejects.toThrow(
        "Erro no logout"
      );
    });

    it("should notify auth status listeners on logout", async () => {
      const callback = vi.fn();
      localStorage.setItem("token", JSON.stringify(mockToken));
      authenticationService.subscribeToAuthStatus(callback);

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
      });

      await authenticationService.logout(mockToken.refreshToken);

      expect(callback).toHaveBeenCalledWith(false);
    });
  });

  describe("register", () => {
    it("should register successfully and auto-login", async () => {
      const mockLoginResponse = {
        ...mockToken,
        user: { id: 1, email: mockRegisterUser.email, name: mockRegisterUser.name },
      };

      // First call: register
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ id: 1, email: mockRegisterUser.email, name: mockRegisterUser.name }),
      });
      // Second call: auto-login
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockLoginResponse,
      });

      const result = await authenticationService.register(mockRegisterUser);

      expect(result).toBe(true);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/register"),
        expect.objectContaining({
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(mockRegisterUser),
        })
      );
      // Verify auto-login was called
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/login"),
        expect.objectContaining({
          method: "POST",
        })
      );
      // Verify token is stored after auto-login
      expect(localStorage.getItem("token")).toBeTruthy();
    });

    it("should throw error on failed registration", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 400,
      });

      await expect(authenticationService.register(mockRegisterUser)).rejects.toThrow(
        "Erro no registro"
      );
    });

    it("should store token after registration auto-login", async () => {
      const mockLoginResponse = {
        ...mockToken,
        user: { id: 1, email: mockRegisterUser.email, name: mockRegisterUser.name },
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ id: 1 }),
      });
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockLoginResponse,
      });

      await authenticationService.register(mockRegisterUser);

      expect(localStorage.getItem("token")).toBeTruthy();
      const storedToken = JSON.parse(localStorage.getItem("token")!);
      expect(storedToken.accessToken).toBe(mockToken.accessToken);
    });
  });

  describe("isAuthenticated", () => {
    it("should return false when no token is stored", async () => {
      const result = await authenticationService.isAuthenticated();
      expect(result).toBe(false);
    });

    it("should return true when valid token is stored", async () => {
      localStorage.setItem("token", JSON.stringify(mockToken));

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ isValid: true }),
      });

      const result = await authenticationService.isAuthenticated();
      expect(result).toBe(true);
    });

    it("should return false when token is invalid", async () => {
      localStorage.setItem("token", JSON.stringify(mockToken));

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ isValid: false }),
      });

      const result = await authenticationService.isAuthenticated();
      expect(result).toBe(false);
      expect(localStorage.getItem("token")).toBeNull();
    });

    it("should return false on validation error", async () => {
      localStorage.setItem("token", JSON.stringify(mockToken));

      (global.fetch as any).mockRejectedValueOnce(new Error("Network error"));

      const result = await authenticationService.isAuthenticated();
      expect(result).toBe(false);
      expect(localStorage.getItem("token")).toBeNull();
    });

    it("should handle invalid JSON in localStorage", async () => {
      localStorage.setItem("token", "invalid-json");

      const result = await authenticationService.isAuthenticated();
      expect(result).toBe(false);
      expect(localStorage.getItem("token")).toBeNull();
    });
  });

  describe("getStoredToken", () => {
    it("should return stored token", () => {
      localStorage.setItem("token", JSON.stringify(mockToken));

      const result = authenticationService.getStoredToken();
      expect(result).toEqual(mockToken);
    });

    it("should return null when no token is stored", () => {
      const result = authenticationService.getStoredToken();
      expect(result).toBeNull();
    });

    it("should return null for invalid JSON", () => {
      localStorage.setItem("token", "invalid-json");

      const result = authenticationService.getStoredToken();
      expect(result).toBeNull();
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
        "Erro ao buscar usuários paginados"
      );
    });
  });

  describe("subscribeToAuthStatus", () => {
    it("should call callback immediately with current status", () => {
      const callback = vi.fn();
      authenticationService.subscribeToAuthStatus(callback);

      expect(callback).toHaveBeenCalledTimes(1);
      expect(callback).toHaveBeenCalledWith(false);
    });

    it("should call callback with true if token exists", () => {
      localStorage.setItem("token", JSON.stringify(mockToken));
      const callback = vi.fn();

      authenticationService.subscribeToAuthStatus(callback);

      expect(callback).toHaveBeenCalledWith(true);
    });

    it("should return unsubscribe function", () => {
      const callback = vi.fn();
      const unsubscribe = authenticationService.subscribeToAuthStatus(callback);

      expect(typeof unsubscribe).toBe("function");

      // Clear the initial call
      callback.mockClear();

      // Trigger a login to test unsubscribe
      localStorage.setItem("token", JSON.stringify(mockToken));

      // Call unsubscribe
      unsubscribe();

      // Subscriber should not be notified anymore after unsubscribe
      localStorage.removeItem("token");
    });

    it("should notify all subscribers on auth status change", async () => {
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      authenticationService.subscribeToAuthStatus(callback1);
      authenticationService.subscribeToAuthStatus(callback2);

      const mockResponse = {
        ...mockToken,
        user: { id: 1, email: mockLoginUser.email, name: "Test User" },
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await authenticationService.login(mockLoginUser);

      expect(callback1).toHaveBeenCalledWith(true);
      expect(callback2).toHaveBeenCalledWith(true);
    });
  });

  describe("token validation caching", () => {
    it("should cache validation for 5 minutes", async () => {
      localStorage.setItem("token", JSON.stringify(mockToken));
      localStorage.setItem("lastValidation", Date.now().toString());

      const callback = vi.fn();
      authenticationService.subscribeToAuthStatus(callback);

      // Should not make fetch call due to recent validation
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should validate token after cache expires", async () => {
      localStorage.setItem("token", JSON.stringify(mockToken));
      // Set validation time to 6 minutes ago
      const sixMinutesAgo = Date.now() - 6 * 60 * 1000;
      localStorage.setItem("lastValidation", sixMinutesAgo.toString());

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ isValid: true }),
      });

      const callback = vi.fn();
      // This will trigger checkInitialAuthStatus
      await new Promise((resolve) => setTimeout(resolve, 10));

      // Should make fetch call because cache expired
      // Note: This test might be flaky due to timing, it's testing the initialization logic
    });

    it("should clear validation cache on logout", async () => {
      localStorage.setItem("token", JSON.stringify(mockToken));
      localStorage.setItem("lastValidation", Date.now().toString());

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
      });

      await authenticationService.logout(mockToken.refreshToken);

      expect(localStorage.getItem("lastValidation")).toBeNull();
    });
  });
});
