export default () => ({
  applicationPort: parseInt(process.env.AUTH_SERVICE_APP_PORT) || 3001,
  jwt: {
    secret: process.env.JWT_SECRET,
    expiresIn: process.env.JWT_EXPIRES_IN,
    refreshSecret: process.env.JWT_REFRESH_TOKEN_SECRET,
    refreshExpiresIn: process.env.JWT_REFRESH_TOKEN_EXPIRES_IN,
  },
  session: {
    secret: process.env.SESSION_SECRET,
    expiresIn: process.env.SESSION_EXPIRES_IN,
  },
  database: {
    host: process.env.POSTGRES_HOST,
    port: parseInt(process.env.POSTGRES_PORT) || 5432,
    username: process.env.POSTGRES_USER,
    password: process.env.POSTGRES_PASSWORD,
    database: process.env.POSTGRES_DATABASE_NAME,
  },
});
