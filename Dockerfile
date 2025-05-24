FROM golang:1.24.3-alpine AS build

ENV YQ_VERSION="v4.45.4" \
    PATH=$PATH:/usr/local/bin

RUN apk add --no-cache make curl tar git bash

RUN curl -sfL "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" -o /usr/local/bin/yq \
    && chmod +x /usr/local/bin/yq

WORKDIR /tmp/helga

COPY . .

RUN make

FROM alpine:latest

COPY --from=build /tmp/helga/bin/helga /tmp/bin/helga

ENTRYPOINT [ "/tmp/bin/helga" ]

