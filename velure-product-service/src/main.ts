import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import * as dotenv from 'dotenv';
import { Logger } from '@nestjs/common';

async function bootstrap() {
  const logger = new Logger('Main');
  
  dotenv.config();
  const PORT = process.env.APP_PORT || 3010;

  const app = await NestFactory.create(AppModule);

  app.enableCors({
    origin: '*',
  });

  await app.listen(PORT);
  logger.log(`Product service is running on port ${PORT}`);
}
bootstrap();