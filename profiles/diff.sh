#!/bin/bash

go tool pprof -top -diff_base=./base.pprof ./result.pprof 