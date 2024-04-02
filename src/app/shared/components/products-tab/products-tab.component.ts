import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

import { ProductCardComponent } from '../product-card/product-card.component';
import { PaginationComponent } from '../pagination/pagination.component';

import { Product } from '../../../utils/interfaces/product.interface';

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
  products: Product[] = [];

  // Pagination
  totalProducts: number = 8;
  currentPageIndex: number = 1;
  itemsPerPage: number = 2;
  paginationDisabled: boolean = false;

  // Product card configuration
  isProductCardHoverable: boolean = true;
  isProductCardBorderless: boolean = false;
  isProductCardLoading: boolean = false;
  isProductCardRateDisabled: boolean = true;

  constructor() { }

  ngOnInit() {
      this.products = [
        {
          name: 'Product 1',
          price: 100,
          rating: 4.3,
          disponibility: true,
          quantity_warehouse: 10,
          images: ['https://via.placeholder.com/100'],
          dimensions: {
            height: 10,
            width: 10,
            length: 10,
            weight: 10
          },
          colors: ['red', 'blue'],
          dt_created: new Date(),
          dt_updated: new Date()
        },
        {
          name: 'Product 2',
          price: 200,
          rating: 3.6,
          disponibility: true,
          quantity_warehouse: 20,
          images: ['https://via.placeholder.com/500'],
          dimensions: {
            height: 20,
            width: 20,
            length: 20,
            weight: 20
          },
          colors: ['green', 'yellow'],
          dt_created: new Date(),
          dt_updated: new Date()
        },
        {
          name: 'Product 3',
          price: 300,
          rating: 2.1,
          disponibility: true,
          quantity_warehouse: 30,
          images: ['https://via.placeholder.com/150',
                  'https://via.placeholder.com/150'],
          dimensions: {
            height: 30,
            width: 30,
            length: 30,
            weight: 30
          },
          colors: ['black', 'white'],
          dt_created: new Date(),
          dt_updated: new Date()
        },
        {
          name: 'Product 4',
          price: 400,
          rating: 0.3,
          disponibility: true,
          quantity_warehouse: 40,
          images: ['https://via.placeholder.com/150',
                  'https://via.placeholder.com/150',
                  'https://via.placeholder.com/150'],
          dimensions: {
            height: 40,
            width: 40,
            length: 40,
            weight: 40
          },
          colors: ['purple', 'orange'],
          dt_created: new Date(),
          dt_updated: new Date()
        }
      ];
      
      this.totalProducts = this.products.length;
  }
}
