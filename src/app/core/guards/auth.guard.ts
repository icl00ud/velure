import { Injectable } from '@angular/core';
import { AuthService } from '../services/auth.service';

@Injectable({
    providedIn: 'root'
})
export class AuthGuard {
    isAuthenticated: boolean = false;

    constructor(
        private authService: AuthService,
    ) { }

    ngOnInit(): void {
        this.isAuthenticated = this.authService.isAuthenticated();
    }
}