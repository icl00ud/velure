import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';

import { CarouselComponent } from '../carousel/carousel.component';

import { NzCardModule } from 'ng-zorro-antd/card';
import { NzRateModule } from 'ng-zorro-antd/rate';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzSkeletonModule } from 'ng-zorro-antd/skeleton';

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
  isHoverable: boolean = true;
  borderless: boolean = false;
  isLoading: boolean = true;
  rateDisabled: boolean = true;
  array: any[] = [];

  constructor() {
    this.array = [
      [
        "https://picsum.photos/seed/picsum/320/100",
        "https://picsum.photos/320/100?grayscale",
      ],
      [
        "https://picsum.photos/320/100",
        "https://picsum.photos/320/100",
        "https://picsum.photos/320/100?grayscale",
      ],
      [
        "https://picsum.photos/320/100",
        "https://picsum.photos/320/100"
      ],
      [
        "https://picsum.photos/320/100?grayscale",
      ],
      [
        "https://picsum.photos/320/100",
        "https://picsum.photos/320/100?grayscale",
      ],
      [
        "https://picsum.photos/320/100",
        "https://picsum.photos/320/100",
        "https://picsum.photos/320/100?grayscale",
      ],
      [
        "https://picsum.photos/320/100",
        "https://picsum.photos/320/100"
      ],
      [
        "https://picsum.photos/320/100?grayscale"
      ]
    ];
  }

  ngOnInit() {
    setTimeout(() => {
      this.isLoading = false;
    }, 2500);
  }
}
