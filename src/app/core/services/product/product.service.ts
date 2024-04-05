import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { Product } from '../../../utils/interfaces/product.interface';

import { ConfigService } from '../config/config.service';

@Injectable({
    providedIn: 'root'
})
export class ProductService {
    constructor(
        private http: HttpClient,
        private config: ConfigService
    ) { }

    getProducts(): Observable<Product[]> {
        return this.http.get<Product[]>(this.config.productApiUrl);
    }

    getProductById(id: number): Observable<Product> {
        const url = `${this.config.productApiUrl}/${id}`;
        return this.http.get<Product>(url);
    }

    getProductsByPage(page: number, pageSize: number): Observable<Product[]> {
        const url = `${this.config.productApiUrl}/v1/GetProductsByPage?page=${page}&pageSize=${pageSize}`;
        return this.http.get<Product[]>(url);
    }

    createProduct(product: any): Observable<any> {
        return this.http.post<any>(this.config.productApiUrl, product);
    }

    updateProduct(id: number, product: any): Observable<any> {
        const url = `${this.config.productApiUrl}/${id}`;
        return this.http.put<any>(url, product);
    }

    deleteProduct(id: number): Observable<any> {
        const url = `${this.config.productApiUrl}/${id}`;
        return this.http.delete<any>(url);
    }
}