#!/usr/bin/env gmake -f

GOOPTS=CGO_ENABLED=0
BUILDOPTS=-ldflags="-s -w" -a -gcflags=all=-l -trimpath -buildvcs=false

all: clean build

build:
	${GOOPTS} go build ${BUILDOPTS} -o seed-fortune      ./cmd/seed-fortune
	${GOOPTS} go build ${BUILDOPTS} -o seed-proverb      ./cmd/seed-proverb
	${GOOPTS} go build ${BUILDOPTS} -o seed-friday       ./cmd/seed-friday
	${GOOPTS} go build ${BUILDOPTS} -o aleesa-phrases-go ./cmd/aleesa-phrases-go

clean:
	rm -rf seed-{fortune,proverb,friday} aleesa-phrases-go

upgrade:
	rm -rf vendor
	go get -u -t -tool ./...
	go mod tidy
	go mod vendor

# vim: set ft=make noet ai ts=4 sw=4 sts=4:
