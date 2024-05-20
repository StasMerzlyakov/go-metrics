#!/bin/bash

go fmt ./...

#  go install golang.org/x/tools/cmd/goimports@latest
goimports -w cmd
goimports -w internal
