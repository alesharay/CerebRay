FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install --prefer-offline
ARG VITE_POSTHOG_KEY
ARG VITE_POSTHOG_HOST
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
RUN addgroup -S app && adduser -S app -G app \
    && chown -R app:app /var/cache/nginx /var/log/nginx \
    && sed -i 's/^user  nginx;/user  app;/' /etc/nginx/nginx.conf
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
