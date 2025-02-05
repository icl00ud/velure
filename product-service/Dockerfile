FROM node:18-alpine AS build

WORKDIR /app

COPY package*.json ./
COPY tsconfig.build.json ./

RUN npm install -g @nestjs/cli

RUN npm install

COPY . .

RUN npm run build

FROM node:18-alpine

WORKDIR /app

COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./

CMD ["node", "dist/main"]
