import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  productApiUrl: string = 'http://localhost:3000/product';

  constructor() { }
}