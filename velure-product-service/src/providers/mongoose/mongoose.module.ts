
import { Module } from '@nestjs/common';
import { databaseProviders } from './mongoose.provider';

@Module({
  providers: [...databaseProviders],
  exports: [...databaseProviders],
})
export class MoongoseModule {}