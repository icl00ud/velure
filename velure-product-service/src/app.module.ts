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
        url: 'redis://localhost:6379',
        retryStrategy: (times) => {
          const maxRetries = 5;
          if (times >= maxRetries) {
            return null;
          }
          return Math.min(times * 50, 2000);
        },
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