import { CommonModule } from '@angular/common';

import { Component, OnInit } from '@angular/core';
import { CartService } from '../../services/cart.service';
import { CurrencyPipe } from '../../pipes/currency.pipe';
import { ProductCardComponent } from '../../../shared/components/product-card/product-card.component';

@Component({
  selector: 'app-cart',
  templateUrl: './cart.component.html',
  styleUrls: ['./cart.component.less'],
  imports: [
    ProductCardComponent,
    CurrencyPipe,
    CommonModule,
  ],
  standalone: true,
})
export class CartComponent implements OnInit {
  cartItems: any[] = [];
  totalPrice = 0;

  constructor(private cartService: CartService) {}

  ngOnInit(): void {
    this.cartItems = this.cartService.getCartItems();
    this.totalPrice = this.cartService.getTotalPrice();
  }

  removeItem(productId: string): void {
    this.cartService.removeFromCart(productId);
    this.cartItems = this.cartService.getCartItems();
    this.totalPrice = this.cartService.getTotalPrice();
  }

  checkout(): void {
    alert('Compra finalizada!');
    this.cartService.clearCart();
    this.cartItems = [];
    this.totalPrice = 0;
  }
}
