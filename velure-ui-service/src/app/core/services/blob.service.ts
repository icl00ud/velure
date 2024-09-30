import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map, switchMap } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class BlobService {

  constructor(private http: HttpClient) { }

  getBase64FromUrl(url: string): Observable<string> {
    return this.http.get(url, { responseType: 'blob' }).pipe(
      map((blob: Blob) => {
        const reader = new FileReader();
        reader.readAsDataURL(blob);

        return new Observable<string>((observer) => {
          reader.onload = () => {
            const base64String = reader.result as string;
            observer.next(base64String.split(',')[1]);
            observer.complete();
          };
          reader.onerror = (error) => {
            observer.error(error);
          };
        });
      }),
      
      switchMap((base64Observable: Observable<string>) => {
        return base64Observable;
      })
    );
  }
}