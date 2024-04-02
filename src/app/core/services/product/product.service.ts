import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { Product } from '../../../utils/interfaces/product.interface';

import { ConfigService } from '../config/config.service';

@Injectable({
    providedIn: 'root'
})
export class ProductService {
    private apiUrl = 'https://api.example.com/products';

    constructor(
        private http: HttpClient,
        private config: ConfigService
    ) { 

    }

    getProducts(): Observable<Product[]> {
        return this.http.get<Product[]>(this.apiUrl);
    }

    getProductById(id: number): Observable<Product> {
        const url = `${this.apiUrl}/${id}`;
        return this.http.get<Product>(url);
    }

    createProduct(product: any): Observable<any> {
        return this.http.post<any>(this.apiUrl, product);
    }

    updateProduct(id: number, product: any): Observable<any> {
        const url = `${this.apiUrl}/${id}`;
        return this.http.put<any>(url, product);
    }

    deleteProduct(id: number): Observable<any> {
        const url = `${this.apiUrl}/${id}`;
        return this.http.delete<any>(url);
    }
}