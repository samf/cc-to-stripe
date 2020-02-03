FROM gliderlabs/alpine:3.6
COPY cc-to-stripe /go/bin/cc-to-stripe
WORKDIR /go/home
RUN apk add --update ca-certificates
EXPOSE 80 443
ENTRYPOINT ["/go/bin/cc-to-stripe"]
VOLUME ["/autocert"]
