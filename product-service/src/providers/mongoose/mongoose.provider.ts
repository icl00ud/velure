import { Logger } from "@nestjs/common";
import * as mongoose from "mongoose";
import { DATABASE_CONNECTION } from "src/shared/constants";

export const databaseProviders = [
  {
    provide: DATABASE_CONNECTION,
    useFactory: async (): Promise<typeof mongoose> => {
      const logger = new Logger("MongoDB");
      const {
        MONGODB_HOST = "localhost",
        MONGODB_NORMAL_USER = "velure_user",
        MONGODB_NORMAL_PASSWORD = "velure_password",
        MONGODB_DBNAME = "velure_database",
        MONGODB_PORT = "27017",
      } = process.env;

      const auth =
        MONGODB_NORMAL_USER && MONGODB_NORMAL_PASSWORD
          ? `${encodeURIComponent(MONGODB_NORMAL_USER)}:${encodeURIComponent(MONGODB_NORMAL_PASSWORD)}@`
          : "";
      const uri = `mongodb://${auth}${MONGODB_HOST}:${MONGODB_PORT}/${MONGODB_DBNAME}`;

      try {
        logger.log(`Connecting to MongoDB at ${MONGODB_HOST}:${MONGODB_PORT}`);
        await mongoose.connect(uri, {});
        logger.log("Successfully connected to MongoDB");
        return mongoose;
      } catch (error) {
        logger.error("Failed to connect to MongoDB", error);
        throw error;
      }
    },
  },
];
