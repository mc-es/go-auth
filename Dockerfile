FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY .app-config.yml /app/.app-config.yml
COPY go-auth /usr/bin/app

ENTRYPOINT ["/usr/bin/app"]
