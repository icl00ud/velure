import { Injectable } from '@angular/core';
import { CanActivate, Router } from '@angular/router';
import { Observable, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { AuthenticationService } from '../services/authentication.service';

@Injectable({
    providedIn: 'root'
})
export class AuthGuard implements CanActivate {

    constructor(
        private authService: AuthenticationService,
        private router: Router
    ) { }

    canActivate(): Observable<boolean> {
        debugger
        return this.authService.isAuthenticated().pipe(
            map(isAuthenticated => {
                if (!isAuthenticated) {
                    this.router.navigate(['/login']);
                    return false;
                }
                return true;
            }),
            catchError(() => {
                this.router.navigate(['/login']);
                return of(false);
            })
        );
    }
}