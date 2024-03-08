import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

import { NzLayoutModule } from 'ng-zorro-antd/layout';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzMenuModule } from 'ng-zorro-antd/menu';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    CommonModule,
    NzLayoutModule,
    NzIconModule,
    NzMenuModule,
  ],
  templateUrl: './header.component.html',
  styleUrl: './header.component.less'
})
export class HeaderComponent {
  hoveringProduct: boolean = false;

  constructor() {}

  hoverProducts(hovering: any) {
    this.hoveringProduct = hovering;
  }
}
