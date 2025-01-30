import {
  Injectable,
  OnModuleInit,
  OnModuleDestroy,
  Logger,
} from '@nestjs/common';
import { PrismaClient } from '@prisma/client';

@Injectable()
export class PrismaService implements OnModuleInit, OnModuleDestroy {
  private prisma: PrismaClient;
  private readonly logger = new Logger(PrismaService.name);

  constructor() {
    this.prisma = new PrismaClient({
      datasources: {
        db: {
          url: process.env.POSTGRES_URL_PRIMARY,
        },
      },
    });
  }

  async onModuleInit() {
    try {
      await this.prisma.$connect();
      this.logger.log('Conectado ao banco de dados primário');
    } catch (primaryError) {
      this.logger.error(
        'Erro ao conectar ao banco de dados primário',
        primaryError,
      );

      this.prisma = new PrismaClient({
        datasources: {
          db: {
            url: process.env.POSTGRES_URL_SECONDARY,
          },
        },
      });

      try {
        await this.prisma.$connect();
        this.logger.log('Conectado ao banco de dados de fallback');
      } catch (fallbackError) {
        this.logger.error(
          'Erro ao conectar ao banco de dados de fallback',
          fallbackError,
        );
        throw fallbackError;
      }
    }
  }

  async onModuleDestroy() {
    await this.prisma.$disconnect();
  }

  get client(): PrismaClient {
    return this.prisma;
  }
}
