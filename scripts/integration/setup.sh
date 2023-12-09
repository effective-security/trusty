#!/bin/bash
set -e

if [[ -z "$ITEST_IMAGE_TAG" ]]; then
    export ITEST_IMAGE_TAG=main
else
    echo "ITEST_IMAGE_TAG=$ITEST_IMAGE_TAG"
    # in CI, sometime the host is reused and errors:
    # "Pool overlaps with other one on this address space"
    #echo "y" | docker network prune || echo "unable to prune networks"
fi

docker compose -f docker-compose.tester.yml -p trusty-ci up --force-recreate --abort-on-container-exit --exit-code-from test-runner

EXIT_CODE=$( docker container ls -a -q | xargs docker inspect -f '{{ .Name }} {{ .State.ExitCode }}' | grep trustyci_test-runner_1 | cut -d ' ' -f2 )
echo "Test exit code: $EXIT_CODE"

exit $EXIT_CODE