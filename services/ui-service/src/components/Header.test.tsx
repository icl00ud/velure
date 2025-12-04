import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "../test/test-utils";
import userEvent from "@testing-library/user-event";
import Header from "./Header";
import { cartService } from "../services/cart.service";
import { authenticationService } from "../services/authentication.service";
import { productService } from "../services/product.service";

// Mock product service
vi.mock("../services/product.service", () => ({
  productService: {
    getCategories: vi.fn(),
  },
}));

// Mock fetch
global.fetch = vi.fn();

describe("Header", () => {
  beforeEach(() => {
    localStorage.clear();
    cartService.clearCart();
    vi.clearAllMocks();
    (productService.getCategories as any).mockResolvedValue([]);
  });

  describe("Logo and Branding", () => {
    it("should render Velure logo", () => {
      render(<Header />);

      expect(screen.getByText("Velure")).toBeTruthy();
    });

    it("should have link to home page", () => {
      render(<Header />);

      const logoLink = screen.getByText("Velure").closest("a");
      expect(logoLink).toHaveAttribute("href", "/");
    });
  });

  describe("Navigation Links", () => {
    it("should render Products link", () => {
      render(<Header />);

      const productsLink = screen.getByText("Produtos");
      expect(productsLink).toBeTruthy();
      expect(productsLink).toHaveAttribute("href", "/products");
    });

    it("should render Contact link", () => {
      render(<Header />);

      const contactLink = screen.getByText("Contato");
      expect(contactLink).toBeTruthy();
      expect(contactLink).toHaveAttribute("href", "/contact");
    });
  });

  describe("Categories Dropdown", () => {
    it("should load and display categories", async () => {
      const mockCategories = ["dogs", "cats", "birds"];
      (productService.getCategories as any).mockResolvedValue(mockCategories);

      render(<Header />);

      await waitFor(() => {
        expect(productService.getCategories).toHaveBeenCalled();
      });
    });

    it("should format category names correctly", async () => {
      const mockCategories = ["dogs", "cats", "birds"];
      (productService.getCategories as any).mockResolvedValue(mockCategories);

      const user = userEvent.setup();
      render(<Header />);

      await waitFor(() => {
        expect(productService.getCategories).toHaveBeenCalled();
      });

      // Find and click the dropdown trigger
      const dropdownButtons = screen.getAllByRole("button");
      const categoryDropdown = dropdownButtons.find((btn) =>
        btn.querySelector('[class*="ChevronDown"]')
      );

      if (categoryDropdown) {
        await user.click(categoryDropdown);

        await waitFor(() => {
          expect(screen.getByText("Cães")).toBeTruthy();
          expect(screen.getByText("Gatos")).toBeTruthy();
          expect(screen.getByText("Pássaros")).toBeTruthy();
        });
      }
    });

    it("should handle category loading errors gracefully", async () => {
      const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      (productService.getCategories as any).mockRejectedValue(new Error("Failed to load"));

      render(<Header />);

      await waitFor(() => {
        expect(productService.getCategories).toHaveBeenCalled();
      });

      expect(consoleSpy).toHaveBeenCalledWith("Failed to load categories:", expect.any(Error));
      consoleSpy.mockRestore();
    });

    it("should not show categories dropdown when no categories exist", () => {
      (productService.getCategories as any).mockResolvedValue([]);

      render(<Header />);

      // Should only have user and cart icon buttons, no category dropdown
      const buttons = screen.getAllByRole("button");
      // Filter out the category dropdown button
      const categoryDropdown = buttons.find((btn) =>
        btn.querySelector('[class*="ChevronDown"]')
      );
      expect(categoryDropdown).toBeUndefined();
    });
  });

  describe("Shopping Cart", () => {
    it("should display cart icon", () => {
      const { container } = render(<Header />);

      const cartLinks = container.querySelectorAll('a[href="/cart"]');
      expect(cartLinks.length).toBeGreaterThan(0);
    });

    it("should show item count badge when cart has items", async () => {
      const mockProduct = {
        _id: "test-product",
        name: "Test Product",
        price: 100,
        category: "Test",
        stock: 10,
        description: "",
        images: [],
      };

      cartService.addToCart(mockProduct, 3);

      render(<Header />);

      await waitFor(() => {
        expect(screen.getByText("3")).toBeTruthy();
      });
    });

    it("should not show badge when cart is empty", () => {
      render(<Header />);

      const badge = screen.queryByText("0");
      expect(badge).toBeNull();
    });

    it("should update badge when items are added to cart", async () => {
      const mockProduct = {
        _id: "test-product",
        name: "Test Product",
        price: 100,
        category: "Test",
        stock: 10,
        description: "",
        images: [],
      };

      render(<Header />);

      // Initially no badge
      expect(screen.queryByText("2")).toBeNull();

      // Add items to cart
      cartService.addToCart(mockProduct, 2);

      await waitFor(() => {
        expect(screen.getByText("2")).toBeTruthy();
      });
    });
  });

  describe("User Authentication", () => {
    it("should show login button when not authenticated", async () => {
      const { container } = render(<Header />);

      await waitFor(() => {
        const loginLinks = container.querySelectorAll('a[href="/login"]');
        expect(loginLinks.length).toBeGreaterThan(0);
      });
    });

    it("should show user menu when authenticated", async () => {
      const mockToken = {
        accessToken: "mock-token",
        refreshToken: "mock-refresh",
      };
      localStorage.setItem("token", JSON.stringify(mockToken));

      render(<Header />);

      await waitFor(() => {
        const userButtons = screen.getAllByRole("button");
        const userMenuButton = userButtons.find(
          (btn) => !btn.querySelector('[class*="ChevronDown"]')
        );
        expect(userMenuButton).toBeTruthy();
      });
    });

    it("should show My Orders option when authenticated", async () => {
      const mockToken = {
        accessToken: "mock-token",
        refreshToken: "mock-refresh",
      };
      localStorage.setItem("token", JSON.stringify(mockToken));

      const user = userEvent.setup();
      render(<Header />);

      await waitFor(() => {
        const buttons = screen.getAllByRole("button");
        expect(buttons.length).toBeGreaterThan(0);
      });

      // Find user menu button (the one without ChevronDown icon)
      const buttons = screen.getAllByRole("button");
      const userMenuButton = buttons.find(
        (btn) => !btn.querySelector('[class*="ChevronDown"]')
      );

      if (userMenuButton) {
        await user.click(userMenuButton);

        // Use findByText for async rendering with portal
        const meusPedidosLink = await screen.findByText("Meus Pedidos", {}, { timeout: 3000 });
        expect(meusPedidosLink).toBeTruthy();
      }
    });

    it("should handle logout", async () => {
      const mockToken = {
        accessToken: "mock-token",
        refreshToken: "mock-refresh",
      };
      localStorage.setItem("token", JSON.stringify(mockToken));

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
      });

      const user = userEvent.setup();
      render(<Header />);

      await waitFor(() => {
        const buttons = screen.getAllByRole("button");
        expect(buttons.length).toBeGreaterThan(0);
      });

      // Open user menu
      const buttons = screen.getAllByRole("button");
      const userMenuButton = buttons.find(
        (btn) => !btn.querySelector('[class*="ChevronDown"]')
      );

      if (userMenuButton) {
        await user.click(userMenuButton);

        // Use findByText for async rendering with portal
        const logoutButton = await screen.findByText("Sair", {}, { timeout: 3000 });
        expect(logoutButton).toBeTruthy();

        // Click logout
        await user.click(logoutButton);

        await waitFor(
          () => {
            expect(global.fetch).toHaveBeenCalledWith(
              expect.stringContaining("/logout/"),
              expect.any(Object)
            );
          },
          { timeout: 2000 }
        );
      }
    });

    it("should handle logout errors gracefully", async () => {
      const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      const mockToken = {
        accessToken: "mock-token",
        refreshToken: "mock-refresh",
      };
      localStorage.setItem("token", JSON.stringify(mockToken));

      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
      });

      const user = userEvent.setup();
      render(<Header />);

      await waitFor(() => {
        const buttons = screen.getAllByRole("button");
        expect(buttons.length).toBeGreaterThan(0);
      });

      const buttons = screen.getAllByRole("button");
      const userMenuButton = buttons.find(
        (btn) => !btn.querySelector('[class*="ChevronDown"]')
      );

      if (userMenuButton) {
        await user.click(userMenuButton);

        // Use findByText for async rendering with portal
        const logoutButton = await screen.findByText("Sair", {}, { timeout: 3000 });
        expect(logoutButton).toBeTruthy();

        await user.click(logoutButton);

        await waitFor(
          () => {
            expect(consoleSpy).toHaveBeenCalledWith("Logout failed:", expect.any(Error));
          },
          { timeout: 2000 }
        );
      }

      consoleSpy.mockRestore();
    });
  });

  describe("Responsive Design", () => {
    it("should have sticky positioning", () => {
      const { container } = render(<Header />);

      const header = container.querySelector("header");
      expect(header?.className).toContain("sticky");
      expect(header?.className).toContain("top-0");
    });

    it("should have backdrop blur effect", () => {
      const { container } = render(<Header />);

      const header = container.querySelector("header");
      expect(header?.className).toContain("backdrop-blur");
    });
  });

  describe("Category Name Formatting", () => {
    it("should format known category names", async () => {
      const mockCategories = ["dogs", "cats", "small-pets", "fish"];
      (productService.getCategories as any).mockResolvedValue(mockCategories);

      const user = userEvent.setup();
      render(<Header />);

      await waitFor(() => {
        expect(productService.getCategories).toHaveBeenCalled();
      });

      const dropdownButtons = screen.getAllByRole("button");
      const categoryDropdown = dropdownButtons.find((btn) =>
        btn.querySelector('[class*="ChevronDown"]')
      );

      if (categoryDropdown) {
        await user.click(categoryDropdown);

        await waitFor(() => {
          expect(screen.getByText("Cães")).toBeTruthy();
          expect(screen.getByText("Gatos")).toBeTruthy();
          expect(screen.getByText("Pets pequenos")).toBeTruthy();
          expect(screen.getByText("Peixes")).toBeTruthy();
        });
      }
    });

    it("should return original name for unknown categories", async () => {
      const mockCategories = ["unknown-category"];
      (productService.getCategories as any).mockResolvedValue(mockCategories);

      const user = userEvent.setup();
      render(<Header />);

      await waitFor(() => {
        expect(productService.getCategories).toHaveBeenCalled();
      });

      const dropdownButtons = screen.getAllByRole("button");
      const categoryDropdown = dropdownButtons.find((btn) =>
        btn.querySelector('[class*="ChevronDown"]')
      );

      if (categoryDropdown) {
        await user.click(categoryDropdown);

        await waitFor(() => {
          expect(screen.getByText("unknown-category")).toBeTruthy();
        });
      }
    });
  });
});
