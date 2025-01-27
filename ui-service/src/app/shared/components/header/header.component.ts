import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NzLayoutModule } from 'ng-zorro-antd/layout';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzMenuModule } from 'ng-zorro-antd/menu';
import { NzDropDownModule } from 'ng-zorro-antd/dropdown';
import { TranslateModule } from '@ngx-translate/core';
import { Router, RouterModule } from '@angular/router';
import { AuthenticationService } from '../../../core/services/authentication.service';
import { Token } from '../../../utils/interfaces/token.interface';
import { Subscription } from 'rxjs';

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
  styleUrls: ['./header.component.less']
})
export class HeaderComponent implements OnInit, OnDestroy {
  hoveringProduct: boolean = false;
  isLoggedIn: boolean = true;
  private authSubscription!: Subscription;

  constructor(
    private authService: AuthenticationService,
    private router: Router
  ) { }

  ngOnInit() {
    this.authSubscription = this.authService.getAuthStatus().subscribe((loggedIn) => {
      this.isLoggedIn = loggedIn;
    });

    this.authService.isAuthenticated().subscribe();
  }

  ngOnDestroy() {
    if (this.authSubscription) {
      this.authSubscription.unsubscribe();
    }
  }

  hoverProducts(hovering: boolean) {
    this.hoveringProduct = hovering;
  }

  logout(): void {
    const tokenString = localStorage.getItem('token');
    if (tokenString) {
      const token: Token = JSON.parse(tokenString);
      this.authService.logout(token.refreshToken).subscribe({
        next: () => {
          this.router.navigate(['/login']);
        },
        error: (err) => {
          console.error('Erro ao realizar logout', err);
        }
      });
    }
  }
}
