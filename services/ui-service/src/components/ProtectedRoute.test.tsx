import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { vi } from "vitest";
import { useAuth } from "@/hooks/use-auth";
import { ProtectedRoute } from "./ProtectedRoute";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: vi.fn(),
}));

const mockUseAuth = vi.mocked(useAuth);

describe("ProtectedRoute", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading state while initializing", () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isInitializing: true,
    } as any);

    const { container } = render(
      <MemoryRouter initialEntries={["/protected"]}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <div>Secret content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>
    );

    expect(container.querySelector(".animate-spin")).toBeInTheDocument();
  });

  it("redirects unauthenticated users to login", () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isInitializing: false,
    } as any);

    render(
      <MemoryRouter initialEntries={["/protected"]}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <div>Secret content</div>
              </ProtectedRoute>
            }
          />
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </MemoryRouter>
    );

    expect(screen.getByText("Login Page")).toBeInTheDocument();
  });

  it("shows protected content when authenticated", () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: true,
      isInitializing: false,
    } as any);

    render(
      <MemoryRouter initialEntries={["/protected"]}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <div>Secret content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>
    );

    expect(screen.getByText("Secret content")).toBeInTheDocument();
  });

  it("redirects authenticated users away from public-only routes", () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: true,
      isInitializing: false,
    } as any);

    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Routes>
          <Route
            path="/login"
            element={
              <ProtectedRoute requireAuth={false} redirectTo="/orders">
                <div>Login form</div>
              </ProtectedRoute>
            }
          />
          <Route path="/orders" element={<div>Orders page</div>} />
        </Routes>
      </MemoryRouter>
    );

    expect(screen.getByText("Orders page")).toBeInTheDocument();
  });
});
