# Сборка бинарей с приложением
FROM golang:1.20 AS build
COPY . /src
WORKDIR /src

RUN --mount=type=cache,target=/go make vendor
RUN --mount=type=cache,target=/root/.cache/go-build make build

# Промежуточный образ, на основе которого будет собран финальный
FROM alpine:3.18.2 AS bin-image
WORKDIR /app
RUN --mount=type=cache,target=/var/cache/apk apk add gcompat

# Final image stages
FROM bin-image AS app-image
COPY --from=build /src/bin/app /usr/local/bin/app
CMD ["/usr/local/bin/app"]
