# Etapa de compilação
FROM node:alpine AS build

WORKDIR /usr/src/velure-store-ui

COPY package.json package-lock.json ./

RUN npm install -g @angular/cli && \
    npm install

COPY . .

RUN ng build --configuration=production

# Etapa final
FROM nginx:alpine

RUN rm /etc/nginx/conf.d/default.conf

COPY nginx.conf /etc/nginx/conf.d/

COPY --from=build /usr/src/velure-store-ui/dist/velure-store-ui/browser /usr/share/nginx/html

CMD ["nginx", "-g", "daemon off;"]
