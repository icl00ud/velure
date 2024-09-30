import { Component, Input } from '@angular/core';
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

  @Input() section: string = '';

  // Pagination
  totalProducts: number = 8;
  currentPageIndex: number = 1;
  paginationDisabled: boolean = false;
  @Input() itemsPerPage: number = 0;

  // Product card configuration
  @Input() productCategory: string = 'all';
  @Input() isProductCardHoverable: boolean = true;
  @Input() isProductCardBorderless: boolean = false;
  @Input() isProductCardLoading: boolean = false;
  @Input() isProductCardRateDisabled: boolean = true;

  constructor(
    private readonly productService: ProductService
  ) { }

  ngOnInit() {
    if(this.section === 'home')
      this.productService.getProductsByPage(this.currentPageIndex, this.itemsPerPage).subscribe(products => this.products = products);  
    
    if(this.section === 'category')
      this.productService.getProductsByPageAndCategory(this.currentPageIndex, this.itemsPerPage, this.productCategory).subscribe(products => this.products = products);

    this.productService.getProductsCount().subscribe(count => this.totalProducts = count);
  }

  handlePageChange(event: any): void {
    this.currentPageIndex = event;

    this.productService.getProductsByPage(this.currentPageIndex, this.itemsPerPage).subscribe(products => {
      this.products = products;
    });
  }
}
