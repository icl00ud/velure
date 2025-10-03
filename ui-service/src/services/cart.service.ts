import type { CartItem, Product } from "../types/product.types";

class CartService {
  private cart: CartItem[] = [];
  private readonly localStorageKey = "velure_cart";
  private readonly maxQuantity = 99;
  private cartListeners: Set<(cart: CartItem[]) => void> = new Set();

  constructor() {
    this.loadCartFromLocalStorage();
  }

  subscribeToCart(callback: (cart: CartItem[]) => void): () => void {
    this.cartListeners.add(callback);
    // Enviar estado atual imediatamente
    callback([...this.cart]);
    // Retornar função de unsubscribe
    return () => this.cartListeners.delete(callback);
  }

  private notifyCartChange(): void {
    // Criar uma cópia profunda para evitar problemas de referência
    const cartCopy = this.cart.map((item) => ({
      ...item,
      product: { ...item.product },
    }));
    this.cartListeners.forEach((callback) => callback(cartCopy));
  }

  addToCart(product: any, quantity: number = 1): void {
    // Normalizar o produto para garantir que tenha _id
    const normalizedProduct = {
      ...product,
      _id: product._id || product.id, // Usar _id se existir, senão usar id
    };

    if (!normalizedProduct || !normalizedProduct._id) {
      console.error("Produto inválido:", normalizedProduct);
      return;
    }

    const existingItemIndex = this.cart.findIndex(
      (item) => item.product._id === normalizedProduct._id
    );

    if (existingItemIndex >= 0) {
      this.cart[existingItemIndex].quantity = Math.min(
        this.cart[existingItemIndex].quantity + quantity,
        this.maxQuantity
      );
    } else {
      const newItem: CartItem = {
        product: { ...normalizedProduct }, // Criar cópia do produto normalizado
        quantity: Math.min(quantity, this.maxQuantity),
      };
      this.cart.push(newItem);
    }

    this.saveCartToLocalStorage();
    this.notifyCartChange();
  }

  updateQuantity(productId: string, quantity: number): void {
    if (!productId) return;

    const itemIndex = this.cart.findIndex((item) => {
      const itemId = item.product._id || (item.product as any).id;
      return itemId === productId;
    });

    if (itemIndex >= 0) {
      if (quantity <= 0) {
        this.removeFromCart(productId);
      } else {
        this.cart[itemIndex].quantity = Math.max(1, Math.min(quantity, this.maxQuantity));
        this.saveCartToLocalStorage();
        this.notifyCartChange();
      }
    }
  }

  getCartItems(): CartItem[] {
    return this.cart.map((item) => ({
      ...item,
      product: { ...item.product },
    }));
  }

  removeFromCart(productId: string): void {
    if (!productId) return;

    const initialLength = this.cart.length;
    this.cart = this.cart.filter((item) => {
      const itemId = item.product._id || (item.product as any).id;
      return itemId !== productId;
    });

    if (this.cart.length !== initialLength) {
      this.saveCartToLocalStorage();
      this.notifyCartChange();
    }
  }
  getTotalPrice(): number {
    return this.cart.reduce((total, item) => {
      const price = Number(item.product.price) || 0;
      const quantity = Number(item.quantity) || 0;
      return total + price * quantity;
    }, 0);
  }

  clearCart(): void {
    this.cart = [];
    this.saveCartToLocalStorage();
    this.notifyCartChange();
  }

  getCartItemsCount(): number {
    return this.cart.reduce((total, item) => total + (Number(item.quantity) || 0), 0);
  }

  isInCart(productId: string): boolean {
    if (!productId) return false;
    return this.cart.some((item) => {
      const itemId = item.product._id || (item.product as any).id;
      return itemId === productId;
    });
  }

  getItemQuantity(productId: string): number {
    if (!productId) return 0;
    const item = this.cart.find((item) => {
      const itemId = item.product._id || (item.product as any).id;
      return itemId === productId;
    });
    return item ? item.quantity : 0;
  }

  private saveCartToLocalStorage(): void {
    try {
      const cartData = JSON.stringify(this.cart);
      localStorage.setItem(this.localStorageKey, cartData);
    } catch (error) {
      console.error("Falha ao salvar no LocalStorage:", error);
    }
  }

  private loadCartFromLocalStorage(): void {
    try {
      const cartData = localStorage.getItem(this.localStorageKey);
      if (cartData) {
        const parsedData: CartItem[] = JSON.parse(cartData);
        this.cart = Array.isArray(parsedData)
          ? parsedData.filter((item) => this.isValidCartItem(item))
          : [];
      } else {
        this.cart = [];
      }
    } catch (error) {
      console.error("Falha ao carregar do LocalStorage:", error);
      this.cart = [];
    }
  }

  private isValidCartItem(item: any): item is CartItem {
    return (
      item &&
      typeof item === "object" &&
      item.product &&
      typeof item.product._id === "string" &&
      typeof item.quantity === "number" &&
      item.quantity > 0
    );
  }
}

export const cartService = new CartService();
