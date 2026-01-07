FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata

COPY go-auth /usr/bin/app

ENTRYPOINT ["/usr/bin/app"]
