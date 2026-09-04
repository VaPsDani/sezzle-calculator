# ---------- build ----------
FROM node:20-alpine AS build

WORKDIR /app

# npm ci installs exactly what package-lock.json pins and fails if the two files
# disagree, which is what makes a CI build reproducible. Both files are copied
# first so this layer is cached until the dependencies actually change.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Runs tsc -b before vite build, so a type error fails the image build.
RUN npm run build

# ---------- runtime ----------
FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
