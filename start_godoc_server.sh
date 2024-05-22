#!/bin/bash

# Без GOFLAGS и при наличии vendoer директорииgodoc не стартует.
# для просмотра всех пакетов (включая internal) http://localhost:8080/pkg/?m=all
# ключ -play позволяет запускать examples (должен)
GOFLAGS="-mod=mod" godoc -play -http=:8080

