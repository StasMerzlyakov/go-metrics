#!/bin/bash

../metricstest -test.v -test.run=^TestIteration1$ -binary-path=cmd/server/server
test $? -eq 0 || ( echo "metrics1 error"; exit 1 )

../metricstest -test.v -test.run=^TestIteration2[AB]*$ -source-path=. -agent-binary-path=cmd/agent/agent
test $? -eq 0 || ( echo "metrics2 error"; exit 1 )

../metricstest -test.v -test.run=^TestIteration3[AB]*$ \
            -source-path=. \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server
test $? -eq 0 || ( echo "metrics3 error"; exit 1 )
