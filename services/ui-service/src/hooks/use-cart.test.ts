import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { cartService } from "../services/cart.service";
import type { Product } from "../types/product.types";
import { useCart } from "./use-cart";

describe("useCart", () => {
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
    localStorage.clear();
    cartService.clearCart();
  });

  it("should initialize with empty cart", () => {
    const { result } = renderHook(() => useCart());

    expect(result.current.cartItems).toEqual([]);
    expect(result.current.totalPrice).toBe(0);
    expect(result.current.itemsCount).toBe(0);
  });

  it("should add product to cart", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 2);
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(1);
      expect(result.current.cartItems[0].product._id).toBe(mockProduct._id);
      expect(result.current.cartItems[0].quantity).toBe(2);
    });
  });

  it("should calculate total price correctly", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 2); // 100 * 2 = 200
      result.current.addToCart(mockProduct2, 3); // 200 * 3 = 600
    });

    await waitFor(() => {
      expect(result.current.totalPrice).toBe(800);
    });
  });

  it("should calculate items count correctly", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 2);
      result.current.addToCart(mockProduct2, 3);
    });

    await waitFor(() => {
      expect(result.current.itemsCount).toBe(5);
    });
  });

  it("should remove product from cart", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 1);
      result.current.addToCart(mockProduct2, 1);
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(2);
    });

    act(() => {
      result.current.removeFromCart(mockProduct._id);
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(1);
      expect(result.current.cartItems[0].product._id).toBe(mockProduct2._id);
    });
  });

  it("should update product quantity", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 2);
    });

    await waitFor(() => {
      expect(result.current.cartItems[0].quantity).toBe(2);
    });

    act(() => {
      result.current.updateQuantity(mockProduct._id, 5);
    });

    await waitFor(() => {
      expect(result.current.cartItems[0].quantity).toBe(5);
    });
  });

  it("should clear cart", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 1);
      result.current.addToCart(mockProduct2, 1);
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(2);
    });

    act(() => {
      result.current.clearCart();
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(0);
      expect(result.current.totalPrice).toBe(0);
      expect(result.current.itemsCount).toBe(0);
    });
  });

  it("should check if product is in cart", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 1);
    });

    await waitFor(() => {
      expect(result.current.isInCart(mockProduct._id)).toBe(true);
      expect(result.current.isInCart("non-existent-id")).toBe(false);
    });
  });

  it("should get item quantity", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 5);
    });

    await waitFor(() => {
      expect(result.current.getItemQuantity(mockProduct._id)).toBe(5);
      expect(result.current.getItemQuantity("non-existent-id")).toBe(0);
    });
  });

  it("should handle product with id instead of _id", async () => {
    const { result } = renderHook(() => useCart());
    const productWithId = { ...mockProduct, id: "product-with-id" };
    delete (productWithId as any)._id;

    act(() => {
      result.current.addToCart(productWithId, 1);
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(1);
      expect(result.current.cartItems[0].product._id).toBe("product-with-id");
    });
  });

  it("should not add invalid products", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(null as any, 1);
      result.current.addToCart({ name: "Invalid" } as any, 1);
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(0);
    });
  });

  it("should handle invalid product IDs in operations", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 1);
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(1);
    });

    act(() => {
      result.current.removeFromCart("" as any);
      result.current.removeFromCart(null as any);
      result.current.updateQuantity("", 5);
      result.current.updateQuantity(null as any, 5);
    });

    await waitFor(() => {
      // Cart should remain unchanged
      expect(result.current.cartItems).toHaveLength(1);
      expect(result.current.cartItems[0].quantity).toBe(1);
    });
  });

  it("should update totals when cart items change", async () => {
    const { result } = renderHook(() => useCart());

    act(() => {
      result.current.addToCart(mockProduct, 1);
    });

    await waitFor(() => {
      expect(result.current.totalPrice).toBe(100);
      expect(result.current.itemsCount).toBe(1);
    });

    act(() => {
      result.current.updateQuantity(mockProduct._id, 3);
    });

    await waitFor(() => {
      expect(result.current.totalPrice).toBe(300);
      expect(result.current.itemsCount).toBe(3);
    });
  });

  it("should maintain function references across rerenders", () => {
    const { result, rerender } = renderHook(() => useCart());

    const initialAddToCart = result.current.addToCart;
    const initialRemoveFromCart = result.current.removeFromCart;
    const initialUpdateQuantity = result.current.updateQuantity;
    const initialClearCart = result.current.clearCart;

    rerender();

    // Functions should be memoized (same reference)
    expect(result.current.addToCart).toBe(initialAddToCart);
    expect(result.current.removeFromCart).toBe(initialRemoveFromCart);
    expect(result.current.updateQuantity).toBe(initialUpdateQuantity);
    expect(result.current.clearCart).toBe(initialClearCart);
  });

  it("should sync with cart service subscription", async () => {
    const { result } = renderHook(() => useCart());

    // Directly modify cart service
    act(() => {
      cartService.addToCart(mockProduct, 3);
    });

    await waitFor(() => {
      expect(result.current.cartItems).toHaveLength(1);
      expect(result.current.cartItems[0].product._id).toBe(mockProduct._id);
      expect(result.current.cartItems[0].quantity).toBe(3);
    });
  });

  it("should cleanup subscription on unmount", () => {
    const { unmount } = renderHook(() => useCart());

    // Should not throw error on unmount
    expect(() => unmount()).not.toThrow();
  });
});
