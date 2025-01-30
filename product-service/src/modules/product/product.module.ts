import { Module } from '@nestjs/common';
import { ProductService } from './product.service';
import { ProductController } from './product.controller';
import { ProductRepository } from './product.repository';
import { productsProviders } from './product.providers';
import { MoongoseModule } from 'src/providers/mongoose/mongoose.module';

@Module({
    imports: [MoongoseModule],
    controllers: [ProductController],
    providers: [
        ProductService,
        ProductRepository,
        ...productsProviders
    ]
})
export class ProductModule { }