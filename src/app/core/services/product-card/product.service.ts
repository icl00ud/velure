import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
    providedIn: 'root'
})
export class ProductService {
    private apiUrl = 'https://api.example.com/products';

    constructor(private http: HttpClient) { }

    getProducts(): Observable<any> {
        return this.http.get<any>(this.apiUrl);
    }

    getProductById(id: number): Observable<any> {
        const url = `${this.apiUrl}/${id}`;
        return this.http.get<any>(url);
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