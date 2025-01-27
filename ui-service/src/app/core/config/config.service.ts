import { Injectable } from '@angular/core';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  productServiceApiUrl: string = `${environment.PRODUCT_SERVICE_URL}/product` || 'http://localhost:3010/product';
  authenticationServiceApiUrl: string = `${environment.AUTHENTICATION_SERVICE_URL}/authentication` || 'http://localhost:3020/authentication';

  constructor() { }

  get productServiceUrl(): string {
    return this.productServiceApiUrl;
  }

  get authenticationServiceUrl(): string {
    return this.authenticationServiceApiUrl;
  }

  getProductServiceUrl(): string {
    return this.productServiceApiUrl;
  }

  getAuthenticationServiceUrl(): string {
    return this.authenticationServiceApiUrl;
  }
  
  getUrls(): any {
    return {
      productServiceUrl: this.productServiceUrl,
      authenticationServiceUrl: this.authenticationServiceUrl
    };
  }
}