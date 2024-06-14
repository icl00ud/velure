import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  productServiceApiUrl: string = 'http://localhost:3000/api/products';
  authenticationServiceApiUrl: string = 'http://localhost:3000/api/auth';

  constructor() { }
}