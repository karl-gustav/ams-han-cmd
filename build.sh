#!/bin/bash

set -e

GOOS=linux GOARCH=arm GOARM=6 go build -o mbus.reader
