FROM golang:1.24-alpine
COPY .. /app
WORKDIR /app
RUN go build -o /app/isitstreamablebot .

FROM alpine:3.21
RUN adduser -D -g nonroot nonroot
COPY --chown=nonroot:nonroot --from=0 /app/isitstreamablebot /usr/local/bin/isitstreamablebot
COPY --chown=nonroot:nonroot ../VERSION /VERSION
USER nonroot
EXPOSE 8080/tcp
ENTRYPOINT ["/usr/local/bin/isitstreamablebot"]
