#!/bin/bash

SERVER_PORT=5473
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=/tmp/go-metrics-metricstest.tmp
KEY_FILE=/tmp/key
echo "keyasd" > $KEY_FILE
DATABASE_DSN='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'


../metricstest -test.v -test.run=^TestIteration1$ -binary-path=cmd/server/server
test $? -eq 0 || { echo "incr1 error" ; exit 1; }

../metricstest -test.v -test.run=^TestIteration2[AB]*$ -source-path=. -agent-binary-path=cmd/agent/agent
test $? -eq 0 || { echo "incr2 error" && exit 1; } 

../metricstest -test.v -test.run=^TestIteration3[AB]*$ \
            -source-path=. \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server
test $? -eq 0 || { echo "incr3 error" && exit 1; }

../metricstest -test.v -test.run=^TestIteration4$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -server-port=$SERVER_PORT \
  -source-path=.
test $? -eq 0 || { echo "incr4 error" && exit 1; }


../metricstest -test.v -test.run=^TestIteration5$ \
   -agent-binary-path=cmd/agent/agent \
   -binary-path=cmd/server/server \
   -server-port=$SERVER_PORT \
   -source-path=.
test $? -eq 0 || { echo "incr5 error" && exit 1; }



../metricstest -test.v -test.run=^TestIteration6$ \
   -agent-binary-path=cmd/agent/agent \
   -binary-path=cmd/server/server \
   -server-port=$SERVER_PORT \
   -source-path=.
test $? -eq 0 || { echo "incr6 error" && exit 1; }


../metricstest -test.v -test.run=^TestIteration7$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -server-port=$SERVER_PORT \
  -source-path=.
test $? -eq 0 || { echo "incr7 error" && exit 1; }



../metricstest -test.v -test.run=^TestIteration8$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -server-port=$SERVER_PORT \
  -source-path=.
test $? -eq 0 || { echo "incr8 error" && exit 1; }


../metricstest -test.v -test.run=^TestIteration9$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -file-storage-path=$TEMP_FILE \
            -server-port=$SERVER_PORT \
            -source-path=.
test $? -eq 0 || { echo "incr9 error" && exit 1; }


../metricstest -test.v -test.run=^TestIteration10[AB]$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -database-dsn=$DATABASE_DSN \
  -server-port=$SERVER_PORT \
  -source-path=.
test $? -eq 0 || { echo "incr10 error" && exit 1; }

../metricstest -test.v -test.run=^TestIteration11$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -database-dsn=$DATABASE_DSN \
  -server-port=$SERVER_PORT \
  -source-path=.
test $? -eq 0 || { echo "incr11 error" && exit 1; }


../metricstest -test.v -test.run=^TestIteration12$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -database-dsn=$DATABASE_DSN \
  -server-port=$SERVER_PORT \
  -source-path=.
test $? -eq 0 || { echo "incr12 error" && exit 1; }

../metricstest -test.v -test.run=^TestIteration13$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -database-dsn=$DATABASE_DSN \
  -server-port=$SERVER_PORT \
  -source-path=.
test $? -eq 0 || { echo "incr13 error" && exit 1; }

../metricstest -test.v -test.run=^TestIteration14$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -database-dsn=$DATABASE_DSN \
  -key="${KEY_FILE}" \
  -server-port=$SERVER_PORT \
  -source-path=.
test $? -eq 0 || { echo "incr14 error" && exit 1; }

