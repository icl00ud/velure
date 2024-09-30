import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  productServiceApiUrl: string = 'http://localhost:3010/product';
  authenticationServiceApiUrl: string = 'http://localhost:3020/authentication';

  constructor() { }
}