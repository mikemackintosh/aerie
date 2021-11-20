FROM alpine:3.13
COPY bin/main /main
EXPOSE 8080/tcp

ENTRYPOINT [ "/main" ]
