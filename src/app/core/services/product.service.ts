import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { Product } from '../../utils/interfaces/product.interface';

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
        return this.http.get<Product[]>(this.config.productServiceApiUrl);
    }

    getProductById(id: number): Observable<Product> {
        const url = `${this.config.productServiceApiUrl}/${id}`;
        return this.http.get<Product>(url);
    }

    getProductsByPage(page: number, pageSize: number): Observable<Product[]> {
        const url = `${this.config.productServiceApiUrl}/getProductsByPage?page=${page}&pageSize=${pageSize}`;
        return this.http.get<Product[]>(url);
    }

    getProductsByPageAndCategory(page: number, pageSize: number, productCategory: string): Observable<Product[]> {
        const url = `${this.config.productServiceApiUrl}/getProductsByPage?page=${page}&pageSize=${pageSize}&category=${productCategory}`;
        return this.http.get<Product[]>(url);
    }

    getProductsCount(): Observable<number> {
        const url = `${this.config.productServiceApiUrl}/getProductsCount`;
        return this.http.get<number>(url);
    }

    createProduct(product: any): Observable<any> {
        return this.http.post<any>(this.config.productServiceApiUrl, product);
    }

    updateProduct(id: number, product: any): Observable<any> {
        const url = `${this.config.productServiceApiUrl}/${id}`;
        return this.http.put<any>(url, product);
    }

    deleteProduct(id: number): Observable<any> {
        const url = `${this.config.productServiceApiUrl}/${id}`;
        return this.http.delete<any>(url);
    }
}