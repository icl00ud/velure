import { Component, Input, input } from '@angular/core';
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
  @Input() carouselData: string[] = [];
  @Input() fatherComponent: string = '';
  @Input() enableAutoPlay: boolean = true;
  @Input() enableSwipe: boolean = false;

  enableDots: boolean = true;

  array = [
    "https://picsum.photos/seed/picsum/1920/1080",
    "https://picsum.photos/1920/1080?grayscale",
    "https://picsum.photos/1920/1080"
  ];

  constructor() {}

  ngOnInit() {
    if (this.carouselData.length)
      this.array = this.carouselData;
  }
}
