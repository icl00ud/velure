import { Component } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { TranslateModule } from '@ngx-translate/core';

import { ProductsTabComponent } from "../../../shared/components/products-tab/products-tab.component";

@Component({
    selector: 'app-category-products',
    standalone: true,
    templateUrl: './category-products.component.html',
    styleUrl: './category-products.component.less',
    imports: [
      ProductsTabComponent,
      TranslateModule
    ]
})
export class CategoryProductsComponent {
  categoryTitle: string = '';

  // Products tab settings
  itemsPerPage: number = 6;

  // Product card settings
  enableProductCardBorderless: boolean = false;
  enableProductCardHoverable: boolean = false;
  enableProductCardLoading: boolean = false;
  enableProductCardRateDisabled: boolean = false;

  constructor(
    private activatedRoute: ActivatedRoute,
  ) { }

  ngOnInit(): void {
    this.categoryTitle = this.activatedRoute.snapshot.data['category'].toUpperCase();
  }
}
