FROM oven/bun:canary-alpine AS build

WORKDIR /app

COPY package*.json ./
COPY tsconfig.build.json ./

RUN bun install -g @nestjs/cli

RUN bun install

COPY . .

RUN bun run prisma generate

RUN bun run build

FROM oven/bun:canary-alpine

WORKDIR /app

COPY ./prisma /app/
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./

RUN bun install prisma --production

RUN bun run prisma generate

ENV PATH=/app/node_modules/.bin:$PATH

CMD ["bun", "run", "dist/main"]
