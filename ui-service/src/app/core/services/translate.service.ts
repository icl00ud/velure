import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';

import { TranslateLoader } from '@ngx-translate/core';
import { Observable, catchError } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class TranslateService implements TranslateLoader{

  constructor(private httpClient: HttpClient) { }

  getTranslation(lang: string): Observable<any> {
    const apiAddress = `../../../assets/i18n/${lang}.json`;
  
    return this.httpClient.get(apiAddress).pipe(
      catchError(error => {
        return this.httpClient.get(`../../../assets/i18n/${lang}.json`);
      })
    );
  }
}