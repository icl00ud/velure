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
  @Input() totalItems: number = 1;
  @Input() pageIndex: number = 1;
  @Input() pageSize: number = 2;
  @Input() paginationDisabled: boolean = false;

  constructor() { }

  onPageChange(event: number): void {
    console.log(event);
    console.log(`cliquei na página ${event}`)
    this.pageIndex = event;
    debugger
  }
}
