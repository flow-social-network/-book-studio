FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/book-studio-api ./cmd/api
RUN mkdir -p /var/lib/book-studio/assets

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build --chown=nonroot:nonroot /out/book-studio-api /book-studio-api
COPY --from=build --chown=nonroot:nonroot /var/lib/book-studio/assets /var/lib/book-studio/assets
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/book-studio-api"]
