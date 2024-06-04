import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  productServiceApiUrl: string = 'http://localhost:3000/product';
  authenticationServiceApiUrl: string = 'http://localhost:3001/authentication';

  constructor() { }
}