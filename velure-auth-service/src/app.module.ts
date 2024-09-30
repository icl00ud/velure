import { Module } from '@nestjs/common';
import { AuthenticationModule } from './modules/authentication/authentication.module';
import { ConfigModule } from '@nestjs/config';
import configurationConfig from './config/configuration.config';
import { PrometheusModule } from '@willsoto/nestjs-prometheus';

@Module({
  imports: [
    ConfigModule.forRoot({
      load: [configurationConfig],
      isGlobal: true,
    }),
    AuthenticationModule,
    PrometheusModule.register({
      path: "/authentication/authMetrics"
    })
  ],
  controllers: [],
  providers: [],
})
export class AppModule {}
