FROM alpine

COPY ca-certificates.crt /etc/ssl/certs/
COPY ./laposte-notify-tracking-static /laposte-notify-tracking

ENTRYPOINT ["/laposte-notify-tracking"]
