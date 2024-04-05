import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

import { ProductCardComponent } from '../product-card/product-card.component';
import { PaginationComponent } from '../pagination/pagination.component';

import { Product } from '../../../utils/interfaces/product.interface';

import { ProductService } from '../../../core/services/product.service';

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
  itemsPerPage: number = 8;
  paginationDisabled: boolean = false;

  // Product card configuration
  isProductCardHoverable: boolean = true;
  isProductCardBorderless: boolean = false;
  isProductCardLoading: boolean = false;
  isProductCardRateDisabled: boolean = true;

  constructor(
    private readonly productService: ProductService
  ) { }

  ngOnInit() {
    this.productService.getProductsCount().subscribe(count => this.totalProducts = count);
    this.productService.getProductsByPage(this.currentPageIndex, this.itemsPerPage).subscribe(products => this.products = products);  
  }

  handlePageChange(event: any): void {
    this.currentPageIndex = event;

    this.productService.getProductsByPage(this.currentPageIndex, this.itemsPerPage).subscribe(products => {
      this.products = products;
    });
  }
}
