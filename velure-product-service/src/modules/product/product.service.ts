import { Injectable } from '@nestjs/common';
import { ProductRepository } from './product.repository';

import { Product } from './interfaces/product.interface';
import { CreateProductDto } from './dto/create-product.dto';
import { ReadProductDTO } from './dto/read-product.dto';

@Injectable()
export class ProductService {
    constructor(private readonly repository: ProductRepository) { }

    async getAllProducts(): Promise<ReadProductDTO[]> {
        return await this.repository.getAllProducts();
    }

    async getProductsByName(name: string): Promise<Product[]> {
        return await this.repository.getProductsByName(name);
    }

    async getProductsByPage(page: number, pageSize: number): Promise<ReadProductDTO[]> {
        return await this.repository.getProductsByPage(page, pageSize);
    }

    async getProductsByPageAndCategory(page: number, pageSize: number, productCategory: string): Promise<ReadProductDTO[]> {
        return await this.repository.getProductsByPageAndCategory(page, pageSize, productCategory);
    }

    async getProductsCount() {
        return await this.repository.getProductsCount();
    }

    async createProduct(createProductDto: CreateProductDto): Promise<Product> {
        return await this.repository.createProduct(createProductDto);
    }

    async deleteProductsByName(name: string): Promise<void> {
        await this.repository.deleteProductsByName(name);
    }

    async deleteProductById(id: string): Promise<void> {
        await this.repository.deleteProductById(id);
    }
}