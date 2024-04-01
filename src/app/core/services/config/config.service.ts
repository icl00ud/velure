import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  productApiUrl: string = 'https://localhost:3000';

  constructor() { }
}