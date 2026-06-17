# syntax=docker/dockerfile:1

FROM node:22-alpine AS builder

WORKDIR /build

COPY .npmrc package.json pnpm-lock.yaml ./
RUN npm config set registry https://registry.npmmirror.com && corepack enable && pnpm install --frozen-lockfile

COPY . .
RUN corepack enable && pnpm build

FROM node:22-alpine

RUN addgroup -S netdisk && adduser -S -G netdisk netdisk

WORKDIR /app

COPY --from=builder /build/build ./build
COPY --from=builder /build/package.json ./

ENV HOST=0.0.0.0 PORT=8090

USER netdisk

EXPOSE 3000

CMD ["node", "build"]
