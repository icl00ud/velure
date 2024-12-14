import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ProductModule } from './modules/product/product.module';
import { RedisModule } from '@nestjs-modules/ioredis';
import { PrometheusModule } from '@willsoto/nestjs-prometheus';
import { HealthModule } from './modules/health/health.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
    }),
    HealthModule,
    ProductModule,
    RedisModule.forRootAsync({
      imports: [ConfigModule],
      useFactory: async (configService: ConfigService) => ({
        type: 'single',
        url: `redis://${configService.get<string>('REDIS_HOST')}:6379`,
        retryStrategy: (times) => {
          const maxRetries = 5;
          if (times >= maxRetries) {
            return null;
          }
          return Math.min(times * 50, 2000);
        },
      }),
      inject: [ConfigService],
    }),
    PrometheusModule.register({
      path: "/metrics",
    }),
  ],
  controllers: [],
  providers: [],
})
export class AppModule {}