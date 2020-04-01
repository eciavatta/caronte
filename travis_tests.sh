#!/bin/bash

# wait mongodb server
sleep 10

# run tests
go test -v

RET_CODE=$?

#
if [ $RET_CODE -eq 0 ]; then
    echo "travis_tests_success"
else
    echo "travis_tests_fails"
fi

exit $RET_CODE
