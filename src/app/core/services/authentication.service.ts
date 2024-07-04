import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { ConfigService } from '../config/config.service';
import { ILoginResponse, ILoginUser, IRegisterUser } from '../../utils/interfaces/user.interface';
import { Observable, of, throwError } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';

@Injectable({
    providedIn: 'root'
})
export class AuthenticationService {
    constructor(
        private readonly http: HttpClient,
        private readonly config: ConfigService
    ) { }

    login(user: ILoginUser): Observable<ILoginResponse> {
        return this.http.post<ILoginResponse>(`${this.config.authenticationServiceApiUrl}/login`, user).pipe(
            tap(() => {
                console.log('Login em andamento...');
            }),
            catchError((error) => {
                console.error('Erro no login', error);
                return throwError(error);
            })
        );
    }

    logout(refreshToken: string): Observable<boolean> {
        return this.http.delete<boolean>(`${this.config.authenticationServiceApiUrl}/logout`, { body: { refreshToken } }).pipe(
            tap(() => {
                console.log('Logout em andamento...');
            }),
            catchError((error) => {
                console.error('Erro no logout', error);
                return throwError(error);
            })
        );
    }

    register(user: IRegisterUser): Observable<boolean> {
        return this.http.post<boolean>(`${this.config.authenticationServiceApiUrl}/register`, user).pipe(
            tap(() => {
                console.log('Registro em andamento...');
            }),
            catchError((error) => {
                console.error('Erro no registro', error);
                return throwError(error);
            })
        );
    }

    isAuthenticated(): Observable<boolean> {
        const userData: string | null = localStorage.getItem('token');
        const parsedUserData: ILoginResponse | null = userData ? JSON.parse(userData) : null;
        if (!parsedUserData || !parsedUserData.accessToken)
            return of(false);
        
        return this.validateToken(parsedUserData.accessToken).pipe(
            map(response => response.isValid),
            catchError(() => of(false))
        );
    }

    private validateToken(token: string): Observable<{ isValid: boolean }> {
        return this.http.post<{ isValid: boolean }>(`${this.config.authenticationServiceApiUrl}/validateToken`, { token }).pipe(
            catchError((error) => {
                console.error('Erro na validação do token', error);
                return of({ isValid: false });
            })
        );
    }
}