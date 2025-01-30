import { Body, Controller, Delete, Get, Param, Post, Query } from '@nestjs/common';
import { InjectRedis } from '@nestjs-modules/ioredis';
import Redis from 'ioredis';

import { ProductService } from './product.service';
import { CreateProductDto } from './dto/create-product.dto';
import { ReadProductDTO } from './dto/read-product.dto';
import { Product } from './interfaces/product.interface';

@Controller('product')
export class ProductController {
    constructor(
        private readonly productService: ProductService,
        @InjectRedis() private readonly redisService: Redis
    ) { }

    @Get()
    async getAllProducts() {
        const cacheKey = 'allProducts';

        try {
            const cachedProducts = await this.redisService.get(cacheKey);
            if (cachedProducts)
                return JSON.parse(cachedProducts);

            const products = await this.productService.getAllProducts();
            await this.redisService.set(cacheKey, JSON.stringify(products), 'EX', 3600);

            return products;
        } catch (error) {
            throw error;
        }
    }

    @Get('getProductsByName/:name')
    async getProductsByName(@Param('name') productName: string): Promise<Product[]> {
        try {
            console.log('getProductsByName', productName)
            return this.productService.getProductsByName(productName);
        } catch (error) {
            throw error;
        }
    }

    @Get('getProductsByPage')
    async getProductsByPage(@Query('page') page: number, @Query('pageSize') pageSize: number): Promise<ReadProductDTO[]> {
        if (!page || !pageSize)
            throw new Error('Missing query parameters');

        const cacheKey = `productsPage:${page}:${pageSize}`;

        try {
            // try to get data from cache
            const cachedProducts = await this.redisService.get(cacheKey);
            if (cachedProducts) {
                return JSON.parse(cachedProducts);
            }

            // if cache is empty, get data from database
            const products = await this.productService.getProductsByPage(page, pageSize);

            // set data to cache
            await this.redisService.set(cacheKey, JSON.stringify(products), 'EX', 3600);

            return products;
        } catch (error) {
            throw error;
        }
    }

    @Get('getProductsByPageAndCategory')
    async getProductsByPageAndCategory(@Query('page') page: number, @Query('pageSize') pageSize: number, @Query('category') productCategory: string): Promise<ReadProductDTO[]> {
        if (!page || !pageSize || !productCategory)
            throw new Error('Missing query parameters');

        try {
            return await this.productService.getProductsByPageAndCategory(page, pageSize, productCategory);
        } catch (error) {
            throw error;
        }
    }

    @Get('getProductsCount')
    async getProductsCount() {
        try {
            return await this.productService.getProductsCount();
        } catch (error) {
            throw error;
        }
    }

    @Post()
    async createProduct(@Body() createProductDto: CreateProductDto): Promise<CreateProductDto> {
        try {
            const result = await this.productService.createProduct(createProductDto);
            await this.redisService.del('allProducts');
            return result;
        } catch (error) {
            throw error;
        }
    }

    @Delete('deleteProductsByName/:name')
    async deleteProductsByName(@Param('name') productName: string): Promise<void> {
        try {
            await this.productService.deleteProductsByName(productName);
            await this.redisService.del('allProducts');
        } catch (error) {
            throw error;
        }
    }

    @Delete('deleteProductById/:id')
    async deleteProductById(@Param('id') productId: string): Promise<void> {
        try {
            await this.productService.deleteProductById(productId);
            await this.redisService.del('allProducts');
        } catch (error) {
            throw error;
        }
    }
}