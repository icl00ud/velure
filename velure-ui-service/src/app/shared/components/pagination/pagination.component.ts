import { Component, EventEmitter, Input, Output } from '@angular/core';

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
  @Input() totalItems: number = 0;
  @Input() pageIndex: number = 0;
  @Input() pageSize: number = 0;
  @Input() paginationDisabled: boolean = false;

  @Output() pageIndexChange: EventEmitter<number> = new EventEmitter<number>();

  constructor() { }

  onPageChange(event: number): void {
    this.pageIndexChange.emit(event);
  }
}
