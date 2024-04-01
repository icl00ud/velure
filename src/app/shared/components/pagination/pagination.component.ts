import { Component, Input } from '@angular/core';

import { NzPaginationModule } from 'ng-zorro-antd/pagination';

@Component({
  selector: 'app-pagination',
  standalone: true,
  imports: [
    NzPaginationModule
  ],
  templateUrl: './pagination.component.html',
  styleUrl: './pagination.component.less'
})
export class PaginationComponent {
  @Input() total: number = 39;
  @Input() pageIndex: number = 1;
  @Input() pageSize: number = 10;
  @Input() paginationDisabled: boolean = false;

  constructor() { }
}
