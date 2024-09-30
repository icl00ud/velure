import { Module } from '@nestjs/common';
import { ProductModule } from './modules/product/product.module';
import { RedisModule } from '@nestjs-modules/ioredis';
import { PrometheusModule } from '@willsoto/nestjs-prometheus';

@Module({
  imports: [
    ProductModule,
    RedisModule.forRootAsync({
      useFactory: () => ({
        type: 'single',
        url: process.env.REDIS_URL || 'redis://localhost:6379',
      }),
    }),
    PrometheusModule.register({
      path: "/product/productMetrics",
    }),
  ],
  controllers: [],
  providers: [],
})
export class AppModule {}