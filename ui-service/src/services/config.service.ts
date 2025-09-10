import { environment } from '../config/environment';

export class ConfigService {
  private readonly productServiceApiUrl: string;
  private readonly authenticationServiceApiUrl: string;

  constructor() {
    this.productServiceApiUrl = `${environment.PRODUCT_SERVICE_URL}/product`;
    this.authenticationServiceApiUrl = `${environment.AUTHENTICATION_SERVICE_URL}/authentication`;
  }

  get productServiceUrl(): string {
    return this.productServiceApiUrl;
  }

  get authenticationServiceUrl(): string {
    return this.authenticationServiceApiUrl;
  }

  getUrls() {
    return {
      productServiceUrl: this.productServiceUrl,
      authenticationServiceUrl: this.authenticationServiceUrl
    };
  }
}

export const configService = new ConfigService();
