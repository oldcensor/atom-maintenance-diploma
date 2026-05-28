# 1. Этап сборки (build stage)
FROM node:16-alpine AS build
WORKDIR /app

# Копируем package.json и package-lock.json
COPY package*.json ./
RUN npm install

# Копируем остальной код и собираем
COPY . .
RUN npm run build

# 2. Этап продакшена (run stage)
FROM nginx:alpine

# Копируем пользовательскую конфигурацию Nginx в контейнер
COPY default.conf /etc/nginx/conf.d/default.conf

# Копируем собранную сборку из первого этапа
COPY --from=build /app/build /usr/share/nginx/html

# По умолчанию Nginx слушает на 80 порту
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
