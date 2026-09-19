export default () => ({
  port: parseInt(process.env.PORT ?? "4000", 10),
  database: {
    url: process.env.DATABASE_URL ?? "postgresql://dashboard:dashboard@localhost:5432/dashboard",
  },
  cors: {
    origin: process.env.CORS_ORIGIN ?? "http://localhost:3000",
  },
  jwt: {
    secret: process.env.JWT_SECRET ?? "change-me",
  },
});
