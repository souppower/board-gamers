#!/bin/sh -eux

cd `dirname $0`

go get -u golang.org/x/tools/cmd/goimports
go get -u github.com/golang/lint/golint

go get -u github.com/constabulary/gb/...
go get -u code.palmstonegames.com/gb-gae

gb vendor restore