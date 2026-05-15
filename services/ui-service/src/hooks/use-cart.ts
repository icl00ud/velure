import { useCallback, useEffect, useState } from "react";
import { cartService } from "../services/cart.service";
import type { CartItem } from "../types/product.types";

export function useCart() {
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [totalPrice, setTotalPrice] = useState<number>(0);
  const [itemsCount, setItemsCount] = useState<number>(0);

  // Recalculate totals whenever cart items change.
  // biome-ignore lint/correctness/useExhaustiveDependencies: Cart totals are derived from the cart service after cartItems changes.
  useEffect(() => {
    setTotalPrice(cartService.getTotalPrice());
    setItemsCount(cartService.getCartItemsCount());
  }, [cartItems]);

  useEffect(() => {
    const unsubscribe = cartService.subscribeToCart((cart) => {
      setCartItems(cart);
    });

    return unsubscribe;
  }, []);

  const addToCart = useCallback((product: any, quantity: number = 1) => {
    // Validar produto antes de normalizar
    if (!product || (!product._id && !product.id)) {
      console.error("Produto inválido:", product);
      return;
    }

    // Normalizar o produto para garantir que tenha _id
    const normalizedProduct = {
      ...product,
      _id: product._id || product.id, // Usar _id se existir, senão usar id
    };

    cartService.addToCart(normalizedProduct, quantity);
  }, []);

  const removeFromCart = useCallback((productId: string) => {
    if (!productId) return;
    cartService.removeFromCart(productId);
  }, []);

  const updateQuantity = useCallback((productId: string, quantity: number) => {
    if (!productId) return;
    cartService.updateQuantity(productId, quantity);
  }, []);

  const clearCart = useCallback(() => {
    cartService.clearCart();
  }, []);

  const isInCart = useCallback((productId: string): boolean => {
    if (!productId) return false;
    return cartService.isInCart(productId);
  }, []);

  const getItemQuantity = useCallback((productId: string): number => {
    if (!productId) return 0;
    return cartService.getItemQuantity(productId);
  }, []);

  return {
    cartItems,
    totalPrice,
    itemsCount,
    addToCart,
    removeFromCart,
    updateQuantity,
    clearCart,
    isInCart,
    getItemQuantity,
  };
}
