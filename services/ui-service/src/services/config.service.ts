import { environment } from "../config/environment";

export class ConfigService {
  private readonly productServiceApiUrl: string;
  private readonly authenticationServiceApiUrl: string;

  constructor() {
    this.productServiceApiUrl = environment.PRODUCT_SERVICE_URL.startsWith("/")
      ? `${environment.PRODUCT_SERVICE_URL}`
      : `${environment.PRODUCT_SERVICE_URL}/product`;
    const authUrl = environment.AUTHENTICATION_SERVICE_URL.replace(/\/+$/, "");
    const normalizedAuthUrl = authUrl
      .replace(/\/authentication$/, "/api")
      .replace(/\/api\/auth$/, "/api");
    this.authenticationServiceApiUrl = normalizedAuthUrl.endsWith("/api")
      ? normalizedAuthUrl
      : `${normalizedAuthUrl}/api`;
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
      authenticationServiceUrl: this.authenticationServiceUrl,
    };
  }
}

export const configService = new ConfigService();
