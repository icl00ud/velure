import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';

import { HeaderComponent } from "./shared/components/header/header.component";
import { FooterComponent } from './shared/components/footer/footer.component';
import { CarouselComponent } from './shared/components/carousel/carousel.component';
import { ProductCardComponent } from './shared/components/product-card/product-card.component';
import { ProductsTabComponent } from './shared/components/products-tab/products-tab.component';

@Component({
    selector: 'app-root',
    standalone: true,
    templateUrl: './app.component.html',
    styleUrls: ['./app.component.less'],
    imports: [
        CommonModule,
        RouterOutlet,
        HeaderComponent,
        FooterComponent,
        CarouselComponent,
        ProductCardComponent,
        ProductsTabComponent
    ]
})
export class AppComponent {

  constructor () { }
}
