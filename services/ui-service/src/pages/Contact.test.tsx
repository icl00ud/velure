import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "../test/test-utils";
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
    expect(screen.getByText("Get in touch")).toBeTruthy();
    expect(screen.getByLabelText(/name/i)).toBeTruthy();
    expect(screen.getByLabelText(/email/i)).toBeTruthy();
    expect(screen.getByLabelText(/message/i)).toBeTruthy();
    expect(screen.getByRole("button", { name: /send message/i })).toBeTruthy();
  });

  it("should show validation errors for empty fields on submit", async () => {
    const user = userEvent.setup();
    render(<Contact />);

    const submitButton = screen.getByRole("button", { name: /send message/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Name must be at least 2 characters.")).toBeTruthy();
      expect(screen.getByText("Please enter a valid email.")).toBeTruthy();
      expect(screen.getByText("Message must be at least 10 characters.")).toBeTruthy();
    });
  });

  it("should show validation error for invalid email", async () => {
    const user = userEvent.setup();
    render(<Contact />);

    const emailInput = screen.getByLabelText(/email/i);
    await user.type(emailInput, "invalid-email");

    const submitButton = screen.getByRole("button", { name: /send message/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Please enter a valid email.")).toBeTruthy();
    });
  });

  it("should handle valid form submission and redirect to mailto", async () => {
    const user = userEvent.setup();
    render(<Contact />);

    // Fill form with valid data
    await user.type(screen.getByLabelText(/name/i), "John Doe");
    await user.type(screen.getByLabelText(/email/i), "john@example.com");
    await user.type(screen.getByLabelText(/message/i), "This is a test message content.");

    const submitButton = screen.getByRole("button", { name: /send message/i });
    await user.click(submitButton);

    await waitFor(() => {
      const expectedSubject = encodeURIComponent("New contact from Velure site");
      expect(window.location.href).toContain(
        `mailto:israelschroederm@gmail.com?subject=${expectedSubject}`
      );
      expect(window.location.href).toContain("John%20Doe"); // Encoded name
    });

    expect(mockToast).toHaveBeenCalledWith({
      title: "Message ready!",
      description: "Your email client will open to send the message.",
    });
  });
});
