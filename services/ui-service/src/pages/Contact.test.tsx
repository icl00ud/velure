import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "../test/test-utils";
import userEvent from "@testing-library/user-event";
import Contact from "./Contact";

// Mock toast
const mockToast = vi.fn();
vi.mock("@/hooks/use-toast", () => ({
  toast: (...args: any[]) => mockToast(...args),
}));

// Mock dependencies
vi.mock("@/components/Header", () => ({
  default: () => <div data-testid="mock-header">Header</div>,
}));

describe("Contact Page", () => {
  // Setup for window.location.href mock
  const originalLocation = window.location;

  beforeEach(() => {
    // Reset window.location mock
    delete (window as any).location;
    (window as any).location = { ...originalLocation, href: "" };
    vi.clearAllMocks();
  });

  afterEach(() => {
    window.location = originalLocation;
  });

  it("should render the contact form correctly", () => {
    render(<Contact />);

    expect(screen.getByTestId("mock-header")).toBeTruthy();
    expect(screen.getByText("Entre em contato")).toBeTruthy();
    expect(screen.getByLabelText(/nome/i)).toBeTruthy();
    expect(screen.getByLabelText(/e-mail/i)).toBeTruthy();
    expect(screen.getByLabelText(/mensagem/i)).toBeTruthy();
    expect(screen.getByRole("button", { name: /enviar mensagem/i })).toBeTruthy();
  });

  it("should show validation errors for empty fields on submit", async () => {
    const user = userEvent.setup();
    render(<Contact />);

    const submitButton = screen.getByRole("button", { name: /enviar mensagem/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Nome deve ter pelo menos 2 caracteres.")).toBeTruthy();
      expect(screen.getByText("Por favor, insira um email válido.")).toBeTruthy();
      expect(screen.getByText("A mensagem deve ter pelo menos 10 caracteres.")).toBeTruthy();
    });
  });

  it("should show validation error for invalid email", async () => {
    const user = userEvent.setup();
    render(<Contact />);

    const emailInput = screen.getByLabelText(/e-mail/i);
    await user.type(emailInput, "invalid-email");
    
    const submitButton = screen.getByRole("button", { name: /enviar mensagem/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Por favor, insira um email válido.")).toBeTruthy();
    });
  });

  it("should handle valid form submission and redirect to mailto", async () => {
    const user = userEvent.setup();
    render(<Contact />);

    // Fill form with valid data
    await user.type(screen.getByLabelText(/nome/i), "John Doe");
    await user.type(screen.getByLabelText(/e-mail/i), "john@example.com");
    await user.type(screen.getByLabelText(/mensagem/i), "This is a test message content.");

    const submitButton = screen.getByRole("button", { name: /enviar mensagem/i });
    await user.click(submitButton);

    await waitFor(() => {
      const expectedSubject = encodeURIComponent("Novo contato via Site Velure");
      expect(window.location.href).toContain(`mailto:israelschroederm@gmail.com?subject=${expectedSubject}`);
      expect(window.location.href).toContain("John%20Doe"); // Encoded name
    });
    
    expect(mockToast).toHaveBeenCalledWith({
      title: "Mensagem preparada!",
      description: "Seu cliente de email será aberto para enviar a mensagem.",
    });
  });
});
