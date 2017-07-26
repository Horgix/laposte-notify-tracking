build::
	go build -o laposte-notify-tracking

static::
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o laposte-notify-tracking-static

get_cacerts::
	cp -L /etc/ssl/certs/ca-certificates.crt ./ca-certificates.crt
