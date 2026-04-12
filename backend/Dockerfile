FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/yarn.lock* ./
RUN yarn install --frozen-lockfile || yarn install
COPY frontend/ ./
ARG VITE_API_URL=/api/v1
ENV VITE_API_URL=${VITE_API_URL}
RUN yarn build

FROM golang:1.25-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/main.go

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata && mkdir -p /app/frontend/dist
COPY --from=backend-builder /out/server /app/server
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist
ENV PORT=8010
ENV FRONTEND_DIST=/app/frontend/dist
EXPOSE 8010
CMD ["/app/server"]
