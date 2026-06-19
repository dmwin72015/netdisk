# syntax=docker/dockerfile:1.7

ARG APP_VERSION

FROM node:22-alpine AS builder
ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION

WORKDIR /build

RUN corepack enable && npm config set registry https://registry.npmmirror.com

COPY .npmrc package.json pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

COPY src/ ./src/
COPY static/ ./static/
COPY messages/ ./messages/
COPY project.inlang/ ./project.inlang/
COPY svelte.config.js tsconfig.json vite.config.ts ./
RUN --mount=type=cache,target=/build/node_modules/.vite \
    pnpm build

FROM node:22-alpine

RUN corepack enable && npm config set registry https://registry.npmmirror.com

RUN addgroup -S netdisk && adduser -S -G netdisk netdisk

WORKDIR /app

COPY --from=builder --chown=netdisk:netdisk /build/build ./build
COPY --from=builder --chown=netdisk:netdisk /build/package.json /build/pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --prod --frozen-lockfile

ENV HOST=0.0.0.0 PORT=8090

USER netdisk

EXPOSE 3000

CMD ["node", "build"]
