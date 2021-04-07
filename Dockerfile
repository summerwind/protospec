FROM golang:1.16 as build

ARG VERSION
ARG COMMIT

WORKDIR /go/src/protospec

COPY go.mod go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-X main._version=${VERSION} -X main._version=${COMMIT}" -o /go/bin/protospec .

####################

FROM gcr.io/distroless/base-debian10

COPY --from=build /go/bin/protospec /usr/local/bin/protospec

CMD ["/usr/local/bin/protospec"]
