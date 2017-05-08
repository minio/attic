<p align="center">
<img src="https://github.com/minio/s3verify/blob/master/s3v_logo-01.png?raw=true" width="140px">
</p>

# s3verify - A tool to test for Amazon S3 V4 Signature API Compatibility [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/minio/minio?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/minio/minio)](https://goreportcard.com/report/minio/minio) [![Build Status](https://travis-ci.org/minio/s3verify.svg?branch=master)](https://travis-ci.org/minio/s3verify)

s3verify performs a series of API calls against an object storage server and checks the responses for AWS S3 signature version 4 compatibility.

## INSTALLATION
### Prerequisites
- A working Golang Environment. If you do not have a working Golang environment, please follow [Install Golang](https://github.com/minio/minio/blob/master/INSTALLGO.md).
- [Minio Server](https://github.com/minio/minio/blob/master/README.md) (optional--only necessary if you wish to test your own Minio Server for S3 Compatibility.)

### From Source
Currently s3verify is only available to be downloaded from source. 

```sh
$ go get -u github.com/minio/s3verify
```

## CLI USAGE

```sh
$ s3verify [FLAGS]
```

### Flags

``s3verify`` implements the following flags:

```
    --help      -h      Prints the help screen.
    --access    -a      Allows user to input their AWS access key.
    --secret    -s      Allows user to input their AWS secret access key.
    --url       -u      Allows user to input the host URL of the server they wish to test.
    --region    -r      Allows user to change the region of the AWS host they are using. 
                        Defaults to 'us-east-1' for non AWS hosts and us-west-1 for AWS hosts
                        (to prevent propogation issues).
    --verbose   -v      Allows user to trace the HTTP requests and responses sent by s3verify.
    --extended          Allows user to decide whether to test only basic or full API compliance.
    --reuse             Allows user to create a new reusable testing environment or reuse an 
                        existing environment, by providing a unique id for the environment.
    --clean             Allows user to remove all s3verify created objects and buckets. 
    --version           Prints the version.
```

### Environment Variables
``s3verify`` also supports the following environment variables as a replacement for flags. In fact it is recommended that on multiuser systems that env. 
variables be used for security reasons.

The following env. variables can be used to replace their corresponding flags.

```
    S3_ACCESS can be set to YOUR_ACCESS_KEY and replaces --access -a.
    S3_SECRET can be set to YOUR_SECRET_KEY and replaces --secret -s.
    S3_REGION can be set to the region of the AWS host and replaces --region -r.
    S3_URL can be set to the host URL of the server users wish to test and replaces --url -u.
```

## EXAMPLES
Use s3verify to check the AWS S3 V4 compatibility of the Minio test server (https://play.minio.io:9000)

```sh
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECRET_KEY https://play.minio.io:9000 
```

Use s3verify to check the AWS S3 V4 compatibility of the Minio test server with all APIs.

```sh
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECERT_KEY https://play.minio.io:9000 --extended
```

If a test fails you can use the verbose flag (--verbose) to check the request and response formed by the test to see where it failed.

```sh
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECRET_KEY https://play.minio.io:9000 --verbose
```

Setting up and then using a reusable testing environment. 
After testing is finished the environment is still accessible with --reuse my-test.

```sh
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECRET_KEY https://play.minio.io:9000 --reuse my-test
```

Removing a reusable environment.

```sh
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECRET_KEY https://play.minio.io:9000 --clean my-test
```
