#!/bin/bash
set -e

docker-compose -f docker-compose.tester.yml -p trusty-ci up --force-recreate --abort-on-container-exit --exit-code-from test-runner

EXIT_CODE=$( docker container ls -a -q | xargs docker inspect -f '{{ .Name }} {{ .State.ExitCode }}' | grep trustyci_test-runner_1 | cut -d ' ' -f2 )
echo "Test exit code: $EXIT_CODE"

docker-compose --version
exit $EXIT_CODE