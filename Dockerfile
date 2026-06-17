# syntax=docker/dockerfile:1

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/http-anatomy .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/http-anatomy /http-anatomy
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/http-anatomy"]
