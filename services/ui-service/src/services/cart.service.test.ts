import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Product } from "../types/product.types";
import { cartService } from "./cart.service";

describe("CartService", () => {
  // Sample product for testing
  const mockProduct: Product = {
    _id: "test-product-1",
    name: "Test Product",
    description: "A test product",
    price: 100,
    category: "Test Category",
    stock: 10,
    images: [],
  };

  const mockProduct2: Product = {
    _id: "test-product-2",
    name: "Test Product 2",
    description: "Another test product",
    price: 200,
    category: "Test Category",
    stock: 5,
    images: [],
  };

  beforeEach(() => {
    // Clear cart and localStorage before each test
    localStorage.clear();
    cartService.clearCart();
  });

  describe("addToCart", () => {
    it("should add a product to the cart", () => {
      cartService.addToCart(mockProduct, 1);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(1);
      expect(items[0].product._id).toBe(mockProduct._id);
      expect(items[0].quantity).toBe(1);
    });

    it("should add multiple quantities of a product", () => {
      cartService.addToCart(mockProduct, 3);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(1);
      expect(items[0].quantity).toBe(3);
    });

    it("should increment quantity when adding existing product", () => {
      cartService.addToCart(mockProduct, 2);
      cartService.addToCart(mockProduct, 3);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(1);
      expect(items[0].quantity).toBe(5);
    });

    it("should respect max quantity limit", () => {
      cartService.addToCart(mockProduct, 50);
      cartService.addToCart(mockProduct, 60);
      const items = cartService.getCartItems();

      expect(items[0].quantity).toBe(99); // Max quantity
    });

    it("should handle product with id instead of _id", () => {
      const productWithId = { ...mockProduct, id: "product-with-id" };
      delete (productWithId as any)._id;

      cartService.addToCart(productWithId, 1);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(1);
      expect(items[0].product._id).toBe("product-with-id");
    });

    it("should not add invalid products", () => {
      cartService.addToCart(null as any, 1);
      cartService.addToCart({ name: "Invalid" } as any, 1);

      const items = cartService.getCartItems();
      expect(items).toHaveLength(0);
    });

    it("should add multiple different products", () => {
      cartService.addToCart(mockProduct, 2);
      cartService.addToCart(mockProduct2, 3);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(2);
      expect(items[0].product._id).toBe(mockProduct._id);
      expect(items[1].product._id).toBe(mockProduct2._id);
    });
  });

  describe("removeFromCart", () => {
    it("should remove a product from cart", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.removeFromCart(mockProduct._id);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(0);
    });

    it("should only remove the specified product", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.addToCart(mockProduct2, 1);
      cartService.removeFromCart(mockProduct._id);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(1);
      expect(items[0].product._id).toBe(mockProduct2._id);
    });

    it("should handle removing non-existent product", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.removeFromCart("non-existent-id");
      const items = cartService.getCartItems();

      expect(items).toHaveLength(1);
    });

    it("should handle invalid product ID", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.removeFromCart("" as any);
      cartService.removeFromCart(null as any);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(1);
    });
  });

  describe("updateQuantity", () => {
    it("should update product quantity", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.updateQuantity(mockProduct._id, 5);
      const items = cartService.getCartItems();

      expect(items[0].quantity).toBe(5);
    });

    it("should respect minimum quantity of 1", () => {
      cartService.addToCart(mockProduct, 5);
      cartService.updateQuantity(mockProduct._id, 0);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(0);
    });

    it("should respect maximum quantity", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.updateQuantity(mockProduct._id, 150);
      const items = cartService.getCartItems();

      expect(items[0].quantity).toBe(99);
    });

    it("should remove item when quantity is zero or negative", () => {
      cartService.addToCart(mockProduct, 5);
      cartService.updateQuantity(mockProduct._id, -1);
      const items = cartService.getCartItems();

      expect(items).toHaveLength(0);
    });

    it("should handle invalid product ID", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.updateQuantity("", 5);
      cartService.updateQuantity(null as any, 5);
      const items = cartService.getCartItems();

      expect(items[0].quantity).toBe(1);
    });
  });

  describe("getCartItems", () => {
    it("should return a copy of cart items", () => {
      cartService.addToCart(mockProduct, 1);
      const items1 = cartService.getCartItems();
      const items2 = cartService.getCartItems();

      expect(items1).toEqual(items2);
      expect(items1).not.toBe(items2); // Different references
    });

    it("should return empty array when cart is empty", () => {
      const items = cartService.getCartItems();
      expect(items).toEqual([]);
    });
  });

  describe("getTotalPrice", () => {
    it("should calculate total price correctly", () => {
      cartService.addToCart(mockProduct, 2); // 100 * 2 = 200
      cartService.addToCart(mockProduct2, 3); // 200 * 3 = 600
      const total = cartService.getTotalPrice();

      expect(total).toBe(800);
    });

    it("should return 0 for empty cart", () => {
      const total = cartService.getTotalPrice();
      expect(total).toBe(0);
    });

    it("should handle invalid prices gracefully", () => {
      const invalidProduct = { ...mockProduct, price: "invalid" as any };
      cartService.addToCart(invalidProduct, 1);
      const total = cartService.getTotalPrice();

      expect(total).toBe(0);
    });
  });

  describe("clearCart", () => {
    it("should remove all items from cart", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.addToCart(mockProduct2, 1);
      cartService.clearCart();
      const items = cartService.getCartItems();

      expect(items).toHaveLength(0);
    });

    it("should clear localStorage", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.clearCart();

      const storedCart = localStorage.getItem("velure_cart");
      expect(JSON.parse(storedCart || "[]")).toEqual([]);
    });
  });

  describe("getCartItemsCount", () => {
    it("should return total number of items", () => {
      cartService.addToCart(mockProduct, 2);
      cartService.addToCart(mockProduct2, 3);
      const count = cartService.getCartItemsCount();

      expect(count).toBe(5);
    });

    it("should return 0 for empty cart", () => {
      const count = cartService.getCartItemsCount();
      expect(count).toBe(0);
    });
  });

  describe("isInCart", () => {
    it("should return true if product is in cart", () => {
      cartService.addToCart(mockProduct, 1);
      expect(cartService.isInCart(mockProduct._id)).toBe(true);
    });

    it("should return false if product is not in cart", () => {
      expect(cartService.isInCart("non-existent-id")).toBe(false);
    });

    it("should handle invalid product ID", () => {
      expect(cartService.isInCart("")).toBe(false);
      expect(cartService.isInCart(null as any)).toBe(false);
    });
  });

  describe("getItemQuantity", () => {
    it("should return quantity of product in cart", () => {
      cartService.addToCart(mockProduct, 5);
      expect(cartService.getItemQuantity(mockProduct._id)).toBe(5);
    });

    it("should return 0 if product is not in cart", () => {
      expect(cartService.getItemQuantity("non-existent-id")).toBe(0);
    });

    it("should handle invalid product ID", () => {
      expect(cartService.getItemQuantity("")).toBe(0);
      expect(cartService.getItemQuantity(null as any)).toBe(0);
    });
  });

  describe("localStorage persistence", () => {
    it("should save cart to localStorage when adding items", () => {
      cartService.addToCart(mockProduct, 2);
      const storedCart = localStorage.getItem("velure_cart");

      expect(storedCart).toBeTruthy();
      const parsed = JSON.parse(storedCart!);
      expect(parsed).toHaveLength(1);
      expect(parsed[0].product._id).toBe(mockProduct._id);
      expect(parsed[0].quantity).toBe(2);
    });

    it("should save cart to localStorage when updating quantity", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.updateQuantity(mockProduct._id, 5);
      const storedCart = localStorage.getItem("velure_cart");
      const parsed = JSON.parse(storedCart!);

      expect(parsed[0].quantity).toBe(5);
    });

    it("should save cart to localStorage when removing items", () => {
      cartService.addToCart(mockProduct, 1);
      cartService.addToCart(mockProduct2, 1);
      cartService.removeFromCart(mockProduct._id);
      const storedCart = localStorage.getItem("velure_cart");
      const parsed = JSON.parse(storedCart!);

      expect(parsed).toHaveLength(1);
      expect(parsed[0].product._id).toBe(mockProduct2._id);
    });
  });

  describe("cart subscription", () => {
    it("should notify subscribers when cart changes", () => {
      const callback = vi.fn();
      const unsubscribe = cartService.subscribeToCart(callback);

      // Should receive initial state immediately
      expect(callback).toHaveBeenCalledTimes(1);
      expect(callback).toHaveBeenCalledWith([]);

      // Should notify on add
      cartService.addToCart(mockProduct, 1);
      expect(callback).toHaveBeenCalledTimes(2);

      unsubscribe();
    });

    it("should send current cart state to new subscribers", () => {
      cartService.addToCart(mockProduct, 2);
      const callback = vi.fn();
      cartService.subscribeToCart(callback);

      expect(callback).toHaveBeenCalledTimes(1);
      const cartData = callback.mock.calls[0][0];
      expect(cartData).toHaveLength(1);
      expect(cartData[0].product._id).toBe(mockProduct._id);
      expect(cartData[0].quantity).toBe(2);
    });

    it("should stop notifying after unsubscribe", () => {
      const callback = vi.fn();
      const unsubscribe = cartService.subscribeToCart(callback);

      cartService.addToCart(mockProduct, 1);
      expect(callback).toHaveBeenCalledTimes(2);

      unsubscribe();
      cartService.addToCart(mockProduct2, 1);
      expect(callback).toHaveBeenCalledTimes(2); // No new calls
    });

    it("should provide deep copies to subscribers", () => {
      const callback = vi.fn();
      cartService.subscribeToCart(callback);
      cartService.addToCart(mockProduct, 1);

      const cartCopy1 = callback.mock.calls[1][0];
      cartService.addToCart(mockProduct2, 1);
      const cartCopy2 = callback.mock.calls[2][0];

      // Should be different instances
      expect(cartCopy1).not.toBe(cartCopy2);
    });
  });
});
