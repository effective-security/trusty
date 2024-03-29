# trusty

![Build](https://github.com/effective-security/trusty/workflows/Build/badge.svg?branch=main)
[![Coverage Status](https://coveralls.io/repos/github/effective-security/trusty/badge.svg?branch=main)](https://coveralls.io/github/effective-security/trusty?branch=main)

Trusty is a Certification Authority.

## Requirements

1. GoLang 1.21+
1. docker compose

To run or test locally, you need GitHub OAuth secret and a random seed.
Add this to your ~/.profile
Note that these secrets can be random if you don't use a specific feature with GCP or Github.

```.sh
export AWS_ACCESS_KEY_ID=notusedbyemulator
export AWS_SECRET_ACCESS_KEY=notusedbyemulator
export AWS_DEFAULT_REGION=us-west-2
export TRUSTY_JWT_SEED=g...A
```

## Build

* `make all` initializes all dependencies, builds and tests.
* `make proto` generates gRPC protobuf.
* `make build` build the executable
* `make gen_test_certs` generate test certificates
* `make test` run the tests
* `make testshort` runs the tests skipping the end-to-end tests and the code coverage reporting
* `make covtest` runs the tests with end-to-end and the code coverage reporting
* `make coverage` view the code coverage results from the last make test run.
* `make generate` runs go generate to update any code gen'd files (query_console.go in our case)
* `make fmt` runs go fmt on the project.
* `make lint` runs the go linter on the project.

run `make all` once, then run `make build` or `make test` as needed.

First run:

    make all

Subsequent builds:

    make build

Tests:

    make test

Optionally run golang race detector with test targets by setting RACE flag:

    make test RACE=true

Review coverage report:

    make coverage

## Generate protobuf

    make proto

## Local testing

When runnning unit tests, the Unix sockets are used. 
If the test fails, there can be `localhost:{port}` files left on the disk.
To clean up use:
    
    find -name "localhost:*" -delete

## Integration test

    make docker docker-citest

## Debug

Add the launch configuration to .vscode/launch.json:

```.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Server",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "remotePath": "",
            "port": 2345,
            "host": "127.0.0.1",
            "program": "${workspaceRoot}/cmd/trusty",
            "env": {},
            "args": [
                "--log-std",
                "--cfg",
                "${workspaceRoot}/etc/dev/trusty-config.yaml"
            ],
            "showLog": false
        }
    ]
}
```

## gRPC issues

To enable gRPC logs:

```sh
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```

and run commands

## Swagger

    make start-swagger

Open http://localhost:8080

Before runing the above command, make sure trusty is running locally using `bin/trusty` command.

### Client config

    cat ~/.config/trusty/config.yaml

```yaml
---
clients:
  local_cis:
    host: http://localhost:7880
    tls:
      trusted_ca: /tmp/trusty/certs/trusty_root_ca.pem
    request:
      retry_limit: 3
      timeout: 6s
  local_ca:
    host: https://localhost:7892
    tls:
      cert: /tmp/trusty/certs/trusty_client.pem
      key: /tmp/trusty/certs/trusty_client.key
      trusted_ca: /tmp/trusty/certs/trusty_root_ca.pem
    request:
      retry_limit: 3
      timeout: 6s
```
