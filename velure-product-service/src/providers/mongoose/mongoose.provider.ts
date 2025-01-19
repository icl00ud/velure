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

      const auth = MONGODB_USER && MONGODB_PASSWORD
        ? `${encodeURIComponent(MONGODB_USER)}:${encodeURIComponent(MONGODB_PASSWORD)}@`
        : '';
      const uri = `mongodb://${auth}${MONGODB_HOST}:${MONGODB_PORT}/${MONGODB_DBNAME}`;

      try {
        logger.log(`Connecting to MongoDB at ${MONGODB_HOST}:${MONGODB_PORT}`);
        await mongoose.connect(uri, {});
        logger.log('Successfully connected to MongoDB');
        return mongoose;
      } catch (error) {
        logger.error('Failed to connect to MongoDB', error);
        throw error;
      }
    },
  },
];
