FROM golang:alpine as builder

ARG APP=ginadmin
ARG VERSION=0.1.0
ARG RELEASE_TAG=${VERSION}

RUN apk add --no-cache gcc musl-dev sqlite-dev

ENV CGO_CFLAGS "-D_LARGEFILE64_SROUCE"

WORKDIR /go/src/${APP}
COPY . .

RUN go build -ldflags "-w -s" -o ./${APP}

FROM alpine

ARG APP=ginadmin

WORKDIR /go/src/${APP}

COPY --from=builder /go/src/${APP}/${APP} /usr/bin

ENTRYPOINT ["ginadmin","start"]
EXPOSE 8080