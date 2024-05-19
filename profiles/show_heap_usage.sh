#!/bin/bash

go tool pprof -http=":9090" $1
