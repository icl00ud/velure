// cart.service.ts
import { Injectable } from '@angular/core';
import { Product } from '../../utils/interfaces/product.interface';
import { BehaviorSubject, Observable } from 'rxjs';

interface CartItem {
  product: Product;
  quantity: number;
}

@Injectable({
  providedIn: 'root',
})
export class CartService {
  private cart: CartItem[] = [];
  private readonly localStorageKey = 'cart';
  private readonly maxQuantity = 99;

  private cartSubject: BehaviorSubject<CartItem[]> = new BehaviorSubject<CartItem[]>([]);

  constructor() {
    this.loadCartFromLocalStorage();
    this.cartSubject.next(this.cart);
  }

  addToCart(product: Product, quantity: number = 1): void {
    const item = this.cart.find((item) => item.product._id === product._id);
    if (item) {
      item.quantity = Math.min(item.quantity + quantity, this.maxQuantity);
    } else {
      this.cart.push({ product, quantity: Math.min(quantity, this.maxQuantity) });
    }
    this.saveCartToLocalStorage();
    this.cartSubject.next(this.cart);
  }

  updateQuantity(productId: string, quantity: number): void {
    const item = this.cart.find((item) => item.product._id === productId);
    if (item) {
      item.quantity = Math.max(1, Math.min(quantity, this.maxQuantity));
      this.saveCartToLocalStorage();
      this.cartSubject.next(this.cart);
    }
  }

  getCartItems(): Observable<CartItem[]> {
    return this.cartSubject.asObservable();
  }

  removeFromCart(productId: string): void {
    this.cart = this.cart.filter((item) => item.product._id !== productId);
    this.saveCartToLocalStorage();
    this.cartSubject.next(this.cart);
  }

  getTotalPrice(): number {
    return this.cart.reduce(
      (total, item) => total + item.product.price * item.quantity,
      0
    );
  }

  clearCart(): void {
    this.cart = [];
    this.saveCartToLocalStorage();
    this.cartSubject.next(this.cart);
  }

  private saveCartToLocalStorage(): void {
    try {
      localStorage.setItem(this.localStorageKey, JSON.stringify(this.cart));
    } catch (error) {
      console.error('Falha ao salvar no LocalStorage:', error);
    }
  }

  private loadCartFromLocalStorage(): void {
    const cartData = localStorage.getItem(this.localStorageKey);
    if (cartData) {
      try {
        const parsedData: CartItem[] = JSON.parse(cartData);
        this.cart = Array.isArray(parsedData) ? parsedData.filter(item => this.isValidCartItem(item)) : [];
      } catch (error) {
        console.error('Falha ao carregar do LocalStorage:', error);
        this.cart = [];
      }
    }
  }

  private isValidCartItem(item: any): item is CartItem {
    return item?.product && typeof item.product._id === 'string' && typeof item.quantity === 'number';
  }
}
