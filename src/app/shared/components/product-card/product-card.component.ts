import { Component, Input } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';

import { CarouselComponent } from '../carousel/carousel.component';

import { NzCardModule } from 'ng-zorro-antd/card';
import { NzRateModule } from 'ng-zorro-antd/rate';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzSkeletonModule } from 'ng-zorro-antd/skeleton';

import { Product } from '../../../utils/interfaces/product.interface';

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
  @Input() product: Product = {} as Product;
  @Input() isHoverable: boolean = false;
  @Input() borderless: boolean = false;
  @Input() isLoading: boolean = false;
  @Input() rateDisabled: boolean = false;
  @Input() section: string = '';

  enableCarouselSwipe: boolean = true;
  enableCarouselAutoPlay: boolean = false;
  enableCarouselDots: boolean = true;

  constructor() { }

  ngOnInit() {
  }
}
