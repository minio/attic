## Minio XL [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/minio/minio?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Minio XL is a minimal cloud storage server for Petascale Storage. Written in Golang and licensed under [Apache license v2](./LICENSE). Compatible with Amazon S3 APIs. [![Build Status](https://travis-ci.org/minio/minio-xl.svg?branch=master)](https://travis-ci.org/minio/minio-xl)

## Description

This version of the Minio binary is built using ``XL`` distribute erasure code backend. ``XL`` erasure codes each data block with - 8 Data x 8 Parity.  ``XL`` is designed for immutable objects.

## Minio Client

[Minio Client (mc)](https://github.com/minio/mc#minio-client-mc-) provides a modern alternative to Unix commands like ``ls``, ``cat``, ``cp``, ``sync``, and ``diff``. It supports POSIX compatible filesystems and Amazon S3 compatible cloud storage systems. It is entirely written in Golang.

## Amazon S3 Compatible Client Libraries
- [Golang Library](https://github.com/minio/minio-go)
- [Java Library](https://github.com/minio/minio-java)
- [Nodejs Library](https://github.com/minio/minio-js)
- [Python Library](https://github.com/minio/minio-py)

## Server Roadmap
~~~
Storage Backend:
- XL: Erasure coded backend.
 - Status: Standalone mode complete.
Storage Operations:
- Controller:
  - Status: Work in progress.

Storage Management:
- Controller UI:
  - Status: Work in progress.
~~~

### Contribute to Minio Project
Please follow Minio [Contributor's Guide](./CONTRIBUTING.md)

### Jobs
If you think in Lisp or Haskell and hack in go, you would blend right in. Send your github link to callhome@minio.io.


