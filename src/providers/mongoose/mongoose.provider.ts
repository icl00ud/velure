import { Logger } from '@nestjs/common';
import * as mongoose from 'mongoose';
import { DATABASE_CONNECTION } from 'src/shared/constants';

export const databaseProviders = [
  {
    provide: DATABASE_CONNECTION,
    useFactory: async (): Promise<typeof mongoose> => {
      const logger = new Logger('MongoDB');

      const {
        MONGODB_HOST = 'localhost',
        MONGODB_USER = 'velure_user',
        MONGODB_PASSWORD = 'velure_password',
        MONGODB_DBNAME = 'velure_database',
        MONGODB_PORT = '27017',
      } = process.env;

      const createConnectionString = (host: string): string => {
        if (MONGODB_USER && MONGODB_PASSWORD) {
          return `mongodb+srv://${encodeURIComponent(MONGODB_USER)}:${encodeURIComponent(MONGODB_PASSWORD)}@${host}/${MONGODB_DBNAME}`;
        }

        return `mongodb://${host}:${MONGODB_PORT}/${MONGODB_DBNAME}`;
      };

      const primaryUrl = createConnectionString(MONGODB_HOST);

      try {
        logger.warn(`Tentando conectar ao MongoDB em ${MONGODB_HOST}:${MONGODB_PORT}`);
        logger.error(`${primaryUrl}`);
        await mongoose.connect(primaryUrl);
        logger.log('Conectado ao MongoDB com sucesso');
        return mongoose;
      } catch (primaryError) {
        logger.error(`Falha ao conectar ao MongoDB`);
      }
    },
  },
];