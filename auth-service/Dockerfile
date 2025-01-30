# Build stage
FROM node:18-alpine AS build

WORKDIR /app

COPY package*.json ./
COPY tsconfig.build.json ./

RUN npm install -g @nestjs/cli

RUN npm install

COPY . .

RUN npx prisma generate

RUN npm run build

# Prod stage
FROM node:18-alpine

WORKDIR /app

COPY ./prisma /app/
COPY ./.env /app/
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./

RUN npm install prisma --omit=dev

RUN npx prisma generate

ENV PATH=/app/node_modules/.bin:$PATH

CMD ["node", "dist/main"]
