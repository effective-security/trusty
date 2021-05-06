# trusty

![Build](https://github.com/ekspand/trusty/workflows/Build/badge.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/ekspand/trusty/badge.svg?branch=master)](https://coveralls.io/github/ekspand/trusty?branch=master)

Trusty is a Certification Authority.

## Requirements

1. GoLang 1.16+
1. SoftHSM 2.6+

To run or test locally, you need GitHub OAuth secret:

```.sh
export TRUSTY_GITHUB_CLIENT_ID=...
export TRUSTY_GITHUB_CLIENT_SECRET=...
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
                "--std",
                "--cfg",
                "${workspaceRoot}/etc/dev/trusty-config.yaml"
            ],
            "showLog": true
        }
    ]
}
```