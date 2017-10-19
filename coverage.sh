#!/usr/bin/env bash

# exit if anything fails
set -e

function removeCoverageOut () {
    if [ -f coverage.out ]
    then
        echo "REMOVING coverage.out"
        rm coverage.out
    fi
}

# delete previous coverage report, if it exists
removeCoverageOut

# run tests
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# delete the new coverage report, if it exists, so I don't accidentally commit it to the repo somehow
removeCoverageOut