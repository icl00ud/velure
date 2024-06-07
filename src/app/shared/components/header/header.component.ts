import { Component, SimpleChange } from '@angular/core';
import { CommonModule } from '@angular/common';

import { NzLayoutModule } from 'ng-zorro-antd/layout';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzMenuModule } from 'ng-zorro-antd/menu';
import { NzDropDownModule } from 'ng-zorro-antd/dropdown';

import { TranslateModule } from '@ngx-translate/core';
import { Router, RouterModule } from '@angular/router';

import { AuthenticationService } from '../../../core/services/authentication.service';
import { ILoginResponse } from '../../../utils/interfaces/user.interface';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    CommonModule,
    NzLayoutModule,
    NzIconModule,
    NzMenuModule,
    NzDropDownModule,
    TranslateModule,
    RouterModule
  ],
  templateUrl: './header.component.html',
  styleUrl: './header.component.less'
})
export class HeaderComponent {
  hoveringProduct: boolean = false;
  showHeader: boolean = true;
  isLoggedIn: boolean = true;

  constructor(
    private authService: AuthenticationService,
    private router: Router
  ) { }

  ngOnInit() {
    this.authService.isAuthenticated().subscribe((loggedIn) => {
      this.isLoggedIn = loggedIn;

      if (this.isLoggedIn) { 
        this.showHeader = true;
      } else { 
        this.showHeader = false;
      }
    });
  }

  hoverProducts(hovering: any) {
    this.hoveringProduct = hovering;
  }

  logout(): boolean {
    const token: ILoginResponse = JSON.parse(localStorage.getItem('token') ?? '') ?? '';
    if (token.refreshToken) {
      this.authService.logout(token.refreshToken).subscribe((response) => {
        localStorage.removeItem('token');
        this.router.navigate(['/login']);
        return response;
      });
    }

    return false;
  }
}