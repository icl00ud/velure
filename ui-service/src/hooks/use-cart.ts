import { useEffect, useState } from "react";
import { cartService } from "../services/cart.service";
import type { CartItem, Product } from "../types/product.types";

export function useCart() {
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [totalPrice, setTotalPrice] = useState<number>(0);
  const [itemsCount, setItemsCount] = useState<number>(0);

  useEffect(() => {
    const unsubscribe = cartService.subscribeToCart((cart) => {
      setCartItems(cart);
      setTotalPrice(cartService.getTotalPrice());
      setItemsCount(cartService.getCartItemsCount());
    });

    return unsubscribe;
  }, []);

  const addToCart = (product: Product, quantity: number = 1) => {
    cartService.addToCart(product, quantity);
  };

  const removeFromCart = (productId: string) => {
    cartService.removeFromCart(productId);
  };

  const updateQuantity = (productId: string, quantity: number) => {
    cartService.updateQuantity(productId, quantity);
  };

  const clearCart = () => {
    cartService.clearCart();
  };

  const isInCart = (productId: string): boolean => {
    return cartItems.some((item) => item.product._id === productId);
  };

  const getItemQuantity = (productId: string): number => {
    const item = cartItems.find((item) => item.product._id === productId);
    return item ? item.quantity : 0;
  };

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
