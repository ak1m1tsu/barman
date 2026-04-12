FROM alpine:3.21

WORKDIR /app

COPY bot .

ENTRYPOINT ["./bot", "--config", "/app/configs/config.yaml"]
