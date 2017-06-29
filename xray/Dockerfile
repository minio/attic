FROM golang:1.8

RUN apt-get update
RUN apt-get install -y -q cmake make wget curl unzip checkinstall pkg-config yasm x264
RUN apt-get install --no-install-recommends -y -q build-essential ca-certificates git

WORKDIR /go/src/github.com/minio/xray
COPY . /go/src/github.com/minio/xray

RUN make

ENTRYPOINT ["xray"]
