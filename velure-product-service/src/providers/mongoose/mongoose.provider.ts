import { Logger } from '@nestjs/common';
import * as mongoose from 'mongoose';
import { DATABASE_CONNECTION } from 'src/shared/constants';

export const databaseProviders = [
  {
    provide: DATABASE_CONNECTION,
    useFactory: async (): Promise<typeof mongoose> => {
      const logger = new Logger('MongoDB');

      const {
        MONGODB_HOST = process.env.MONGODB_HOST || 'localhost',
        MONGODB_FALLBACK_HOST = process.env.MONGODB_FALLBACK_HOST || 'localhost',
        MONGODB_USER,
        MONGODB_PASSWORD,
        MONGODB_DBNAME,
        MONGODB_PORT = '27017',
      } = process.env;

      const createConnectionString = (host: string): string => {
        if (MONGODB_USER && MONGODB_PASSWORD) {
          return `mongodb://${encodeURIComponent(MONGODB_USER)}:${encodeURIComponent(MONGODB_PASSWORD)}@${host}:${MONGODB_PORT}/${MONGODB_DBNAME}?authSource=admin`;
        }

        return `mongodb://${host}:${MONGODB_PORT}/${MONGODB_DBNAME}`;
      };

      const primaryUrl = createConnectionString(MONGODB_HOST);
      const fallbackUrl = createConnectionString(MONGODB_FALLBACK_HOST);

      try {
        logger.warn(`Tentando conectar ao MongoDB primário em ${MONGODB_HOST}:${MONGODB_PORT}`);
        await mongoose.connect(primaryUrl);
        logger.log('Conectado ao MongoDB primário com sucesso');
        return mongoose;
      } catch (primaryError) {
        logger.error(`Falha ao conectar ao MongoDB primário`);

        try {
          logger.log(`Tentando conectar ao MongoDB de fallback em ${MONGODB_FALLBACK_HOST}:${MONGODB_PORT}`);
          await mongoose.connect(fallbackUrl);
          logger.log('Conectado ao MongoDB de fallback com sucesso');
          return mongoose;
        } catch (fallbackError) {
          logger.error(`Falha ao conectar ao MongoDB de fallback`);
          throw new Error('Não foi possível conectar ao MongoDB (tanto o primário quanto o fallback falharam)');
        }
      }
    },
  },
];
