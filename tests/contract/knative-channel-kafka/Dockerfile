FROM golang:1.10-alpine as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/contract/knative-channel-kafka/

WORKDIR ${SRC_DIR}
COPY .  ${SRC_DIR}

RUN apk add bash dep git
RUN go test -c ./ -o /kafka_channel.test

FROM alpine:3.8

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /kafka_channel.test /kafka_channel.test
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/contract/knative-channel-kafka/licenses /licenses

ENTRYPOINT [ "/kafka_channel.test", "-test.v" ]
