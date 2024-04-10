import { Component } from '@angular/core';
import { CarouselComponent } from '../../components/carousel/carousel.component';
import { ProductsTabComponent } from '../../components/products-tab/products-tab.component';

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

}
