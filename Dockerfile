FROM ubuntu:latest

RUN apt-get update && apt-get install -y ca-certificates curl \
  && update-ca-certificates

ARG APP_NAME
COPY build/bin/${APP_NAME} /usr/local/bin/${APP_NAME}
RUN chmod a+rx /usr/local/bin/${APP_NAME}

ENV APP_NAME=${APP_NAME}
ENTRYPOINT ["/usr/local/bin/${APP_NAME}"]
