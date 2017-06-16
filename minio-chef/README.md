# Minio-Chef Cookbook

[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/minio/minio?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![License](https://img.shields.io/badge/license-Apache_2-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

[Chef](https://chef.io) cookbook to install and do basic configuration of Minio, both [server](https://github.com/minio/minio) and [client](https://github.com/minio/mc).

For server, sets up a basic service with [poise-service](https://github.com/poise/poise-service)

For client, uses chef search to find servers so set up, and seeds credentials for them to `mc` for a specified user.

Both with [custom resources](https://docs.chef.io/custom_resources.html)

Requires `chef 12.5` or newer, or maybe `resource_compat` cookbook

## Basic example

```ruby
minio_server '/srv/minio' do
  service_provider :systemd
end

minio_client_for 'developer'
```
