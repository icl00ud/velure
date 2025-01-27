import { Component } from '@angular/core';

import { CarouselComponent } from '../../../shared/components/carousel/carousel.component';
import { ProductsTabComponent } from '../../../shared/components/products-tab/products-tab.component';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [
    CarouselComponent,
    ProductsTabComponent
  ],
  templateUrl: './home.component.html',
  styleUrl: './home.component.less'
})
export class HomeComponent {
  // Carousel settings
  enableCarouselDots: boolean = false;
  enableCarouselSwipe: boolean = true;

  // Product card settings
  enableProductCardBorderless: boolean = false;
  enableProductCardHoverable: boolean = true;
  enableProductCardLoading: boolean = false;
  enableProductCardRateDisabled: boolean = true;

  // Products tab settings
  itemsPerPage: number = 12;

  constructor() { }
}
