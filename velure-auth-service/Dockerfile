# Estágio de construção
FROM node:alpine AS build

WORKDIR /app

COPY package*.json ./

RUN npm install --production

COPY . .

RUN npm run build

# Estágio de produção
FROM node:alpine

WORKDIR /app

COPY ./prisma /app/
COPY ./.env /app/
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./

RUN npm install prisma --omit=dev

ENV PATH=/app/node_modules/.bin:$PATH