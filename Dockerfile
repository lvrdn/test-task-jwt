FROM golang:1.23 AS dependencies
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download


FROM dependencies AS build
WORKDIR /app
COPY . .
RUN go build -o bin/app cmd/main.go


FROM debian:stable-slim
WORKDIR /app
COPY --from=build /app/bin/app .
COPY /scripts/wait-for-it.sh .
RUN chmod +x /app/app
CMD ["/app/wait-for-it.sh","dbPostgresql:5432","--","/app/app"]