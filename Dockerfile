FROM node:lts-alpine as frontend
LABEL org.opencontainers.image.source https://github.com/astriaorg/seq-faucet

WORKDIR /frontend-build

COPY web/package.json web/yarn.lock ./
RUN yarn install

COPY web ./
RUN yarn build

FROM golang:1.17-alpine as backend

RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /backend-build

COPY go.* ./
RUN go mod download

COPY . .
COPY --from=frontend /frontend-build/dist web/dist

RUN go build -o seq-faucet -ldflags "-s -w"

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=backend /backend-build/seq-faucet /app/seq-faucet

EXPOSE 8080

ENTRYPOINT ["/app/seq-faucet"]
