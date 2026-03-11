#-- 1. Build CSS --
FROM node:20-alpine AS ccs-builder

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY static/css/input.css ./static/css/input.css
COPY views/ ./views/
COPY postcss.config.js ./

RUN npx @tailwindcss/cli \
  -i  ./static/css/input.css \
  -o   ./static/css/output.css

#-- 2. Build Go binary --     
FROM golang:1.23-alpine AS go-builder

WORKDIR /app

#-- Install build dependencies --
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w" \
  -o ./bin/url \
  ./cmd/web

#-- 3. Minimal runtime image --
FROM alpine:3.20

WORKDIR /app

# ca-certificates needed for HTTPS outbound (mailer, etc.)
RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=go-builder /app/bin/url ./url

# Copy views and static assets
COPY --from=go-builder /app/views ./views
COPY --from=go-builder /app/migrations ./migrations
COPY --from=css-builder /app/static/css/output.css ./static/css/output.css
COPY static/ ./static/

EXPOSE 8080

CMD ["./url"]