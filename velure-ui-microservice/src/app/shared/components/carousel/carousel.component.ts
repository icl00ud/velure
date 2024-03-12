import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

import { NzCarouselModule } from 'ng-zorro-antd/carousel';

@Component({
  selector: 'app-carousel',
  standalone: true,
  imports: [
    CommonModule,
    NzCarouselModule
  ],
  templateUrl: './carousel.component.html',
  styleUrl: './carousel.component.less'
})
export class CarouselComponent {
  array = [
    "https://picsum.photos/seed/picsum/1250/400",
    "https://picsum.photos/1250/400?grayscale",
    "https://picsum.photos/1250/400/?blur"
  ];
}
