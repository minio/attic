FROM ubuntu-debootstrap:14.04

MAINTAINER Minio Community

ENV GOLANG_TARBALL go1.5.1.linux-amd64.tar.gz

ENV GOROOT /usr/local/go/
ENV GOPATH /go-workspace
ENV PATH ${GOROOT}/bin:${GOPATH}/bin/:$PATH

ENV MINIOHOME /home/minio-xl
ENV MINIOUSER minio-xl
RUN useradd -m -d $MINIOHOME $MINIOUSER

RUN apt-get update -y && apt-get install -y -q \
		curl \
		git \
		build-essential \
		ca-certificates \
		yasm

RUN curl -O -s https://storage.googleapis.com/golang/${GOLANG_TARBALL} && \
		tar -xzf ${GOLANG_TARBALL} -C ${GOROOT%*go*} && \
		rm ${GOLANG_TARBALL}

ADD . ${GOPATH}/src/github.com/minio/minio-xl

RUN cd ${GOPATH}/src/github.com/minio/minio-xl && \
		make

RUN apt-get remove -y build-essential curl git && \
        apt-get -y autoremove && \
        rm -rf /var/lib/apt/lists/*

USER minio-xl

EXPOSE 9000 9001

CMD ["sh", "-c", "${GOPATH}/bin/minio-xl server"]
