FROM golang:1.21.6-alpine3.19 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/app

RUN addgroup gouser &&  \
    adduser --ingroup gouser --uid 19998 --shell /bin/false gouser --disabled-password && \
    cat /etc/passwd | grep gouser > /etc/passwd_gouser

FROM scratch as runner

COPY --from=builder /etc/passwd_gouser /etc/passwd
COPY --from=builder /go/bin/app /go/bin/app
USER gouser

ENV DB_HOST=
ENV DB_PORT=
ENV DB_USER=
ENV DB_PASSWORD=
ENV DB_DATABASE=
ENV DB_MIN_CONNECTIONS=
ENV DB_MAX_CONNECTIONS=
ENV DB_MAX_CONN_IDLE_TIME=
ENV DB_MAX_CONN_LIFETIME=
ENV GO_ENV=production
ENV GIN_MODE=release

ENTRYPOINT ["/go/bin/app"]
