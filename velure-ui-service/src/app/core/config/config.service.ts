import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  productServiceApiUrl: string = `${import.meta.env.NG_APP_PRODUCT_SERVICE_URL}/product` || 'http://localhost:3010/product';
  authenticationServiceApiUrl: string = `${import.meta.env.NG_APP_AUTHENTICATION_SERVICE_URL}/authentication` || 'http://localhost:3020/authentication';

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