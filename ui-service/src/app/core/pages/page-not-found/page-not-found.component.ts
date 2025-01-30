import { Component } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

import { NzResultModule } from 'ng-zorro-antd/result';

@Component({
  selector: 'app-page-not-found',
  standalone: true,
  imports: [
    NzResultModule
  ],
  templateUrl: './page-not-found.component.html',
  styleUrl: './page-not-found.component.less'
})
export class PageNotFoundComponent {
  subTitle: string = "";

  constructor(private translateService: TranslateService) {}

  ngOnInit() {
    this.translateService.get('GLOBAL.404').subscribe((res: string) => {
      this.subTitle = res;
    });
  }
}
