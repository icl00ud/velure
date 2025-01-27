import { Inject, Injectable } from '@nestjs/common';
import { Product } from './interfaces/product.interface';
import { Model } from 'mongoose';
import { CreateProductDto } from './dto/create-product.dto';
import { ReadProductDTO } from './dto/read-product.dto';
import { PRODUCT_MODEL } from '../../shared/constants';

@Injectable()
export class ProductRepository {
    constructor(
        @Inject(PRODUCT_MODEL)
        private readonly productModel: Model<Product>
    ) { }

    async getAllProducts(): Promise<ReadProductDTO[]> {
        return await this.productModel.find().exec();
    }

    async getProductsByName(name: string): Promise<Product[]> {
        return await this.productModel.find({ name: name }).exec();
    }

    async getProductsByPage(page: number, pageSize: number): Promise<Product[]> {
        return await this.productModel.find().skip((page - 1) * pageSize).limit(pageSize).exec();
    }

    async getProductsByPageAndCategory(page: number, pageSize: number, productCategory: string): Promise<Product[]> {
        return await this.productModel.find({ category: productCategory }).skip((page - 1) * pageSize).limit(pageSize).exec();
    }

    async getProductsCount() {
        return await this.productModel.countDocuments().exec();
    }

    async createProduct(createProductDto: CreateProductDto): Promise<Product> {
        return await this.productModel.create(createProductDto);
    }

    async deleteProductsByName(name: string): Promise<void> {
        await this.productModel.deleteMany({ name });
    }

    async deleteProductById(id: string): Promise<void> {
        await this.productModel.deleteOne({ _id: id });
    }
}