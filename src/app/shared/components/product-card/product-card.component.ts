import { Component, Input } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';

import { CarouselComponent } from '../carousel/carousel.component';

import { NzCardModule } from 'ng-zorro-antd/card';
import { NzRateModule } from 'ng-zorro-antd/rate';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzSkeletonModule } from 'ng-zorro-antd/skeleton';

import { ProductService } from '../../../core/services/product/product.service';
import { Product } from '../../interface/product.interface';

@Component({
  selector: 'app-product-card',
  standalone: true,
  imports: [
    CarouselComponent,
    CommonModule,
    FormsModule,
    NzCardModule,
    NzRateModule,
    NzGridModule,
    NzSkeletonModule
  ],
  templateUrl: './product-card.component.html',
  styleUrl: './product-card.component.less'
})
export class ProductCardComponent {
  @Input() productData: Product[] = [];
  @Input() isHoverable: boolean = true;
  @Input() borderless: boolean = false;
  @Input() isLoading: boolean = false;
  @Input() rateDisabled: boolean = true;

  constructor(
    private productService: ProductService
  ) { 
    this.productService.getProducts().subscribe((products: Product[]) => {
      this.productData = products;
    });
  }

  ngOnInit() {
  }
}
