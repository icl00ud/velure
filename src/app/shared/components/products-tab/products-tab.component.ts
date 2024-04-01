import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

import { ProductCardComponent } from '../product-card/product-card.component';
import { PaginationComponent } from '../pagination/pagination.component';

@Component({
  selector: 'app-products-tab',
  standalone: true,
  imports: [
    CommonModule,
    ProductCardComponent,
    PaginationComponent
  ],
  templateUrl: './products-tab.component.html',
  styleUrl: './products-tab.component.less'
})
export class ProductsTabComponent {
  products: any[] = [];

  constructor() { }

  ngOnInit() {
    setTimeout(() => {
      this.products = [];
    }, 2500);
  }
}
