import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ConfigService } from '../config/config.service';
import { ILoginResponse, ILoginUser, IRegisterUser } from '../../utils/interfaces/user.interface';
import { Observable, BehaviorSubject, of, throwError } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';
import { Token } from '@angular/compiler';

@Injectable({
  providedIn: 'root'
})
export class AuthenticationService {
  private authStatus = new BehaviorSubject<boolean>(this.hasToken());

  constructor(
    private readonly http: HttpClient,
    private readonly config: ConfigService
  ) { }

  private hasToken(): boolean {
    const tokenString = localStorage.getItem('token');
    return !!tokenString;
  }

  getAuthStatus(): Observable<boolean> {
    return this.authStatus.asObservable();
  }

  login(user: ILoginUser): Observable<ILoginResponse> {
    return this.http.post<ILoginResponse>(`${this.config.authenticationServiceApiUrl}/login`, user).pipe(
      tap((response) => {
        localStorage.setItem('token', JSON.stringify(response));
        this.authStatus.next(true);
        console.log('Login realizado com sucesso.');
      }),
      catchError((error) => {
        console.error('Erro no login', error);
        return throwError(error);
      })
    );
  }

  logout(refreshToken: string): Observable<boolean> {
    return this.http.delete<boolean>(`${this.config.authenticationServiceApiUrl}/logout/${refreshToken}`).pipe(
      tap(() => {
        localStorage.removeItem('token');
        this.authStatus.next(false);
        console.log('Logout realizado com sucesso.');
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
        console.log('Registro realizado com sucesso.');
      }),
      catchError((error) => {
        console.error('Erro no registro', error);
        return throwError(error);
      })
    );
  }

  isAuthenticated(): Observable<boolean> {
    const tokenString = localStorage.getItem('token');
    if (!tokenString) {
      return of(false);
    }

    const token: Token = JSON.parse(tokenString);
    return this.validateToken(token).pipe(
      map(response => response.isValid),
      tap(isValid => this.authStatus.next(isValid)),
      catchError(() => {
        this.authStatus.next(false);
        return of(false);
      })
    );
  }

  private validateToken(token: Token): Observable<{ isValid: boolean }> {
    return this.http.post<{ isValid: boolean }>(`${this.config.authenticationServiceApiUrl}/validateToken`, { token }).pipe(
      catchError((error) => {
        console.error('Erro na validação do token', error);
        return of({ isValid: false });
      })
    );
  }
}
