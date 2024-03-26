# Сборка бинарей с приложением
FROM golang:1.20 AS build
COPY . /src
WORKDIR /src

ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/go go mod download
RUN --mount=type=cache,target=/root/.cache/go-build go build -o bin/app .

# Final image stages
FROM alpine:3.19 AS app
COPY --from=build /src/bin/app /usr/local/bin/app
CMD ["/usr/local/bin/app"]
