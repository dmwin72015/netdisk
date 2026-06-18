# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS builder

WORKDIR /build

RUN corepack enable && npm config set registry https://registry.npmmirror.com

COPY .npmrc package.json pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

COPY . .
ARG GIT_SHA=""
ENV VITE_GIT_SHA=${GIT_SHA}
RUN --mount=type=cache,target=/build/node_modules/.vite \
    pnpm build

FROM node:22-alpine

RUN addgroup -S netdisk && adduser -S -G netdisk netdisk

WORKDIR /app

COPY --from=builder --chown=netdisk:netdisk /build/build ./build
COPY --from=builder --chown=netdisk:netdisk /build/package.json ./

ENV HOST=0.0.0.0 PORT=8090

USER netdisk

EXPOSE 3000

CMD ["node", "build"]
