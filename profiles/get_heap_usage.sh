#!/bin/bash

curl http://127.0.0.1:8080/debug/pprof/heap > $1
