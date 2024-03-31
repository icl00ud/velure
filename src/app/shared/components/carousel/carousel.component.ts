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
  enableSwipe: boolean = false;
  enableDots: boolean = true;
  enableAutoPlay: boolean = true;

  array = [
    "https://picsum.photos/seed/picsum/1920/1080",
    "https://picsum.photos/1920/1080?grayscale",
    "https://picsum.photos/1920/1080"
  ];
}
