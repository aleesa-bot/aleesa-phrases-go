#!/usr/bin/env gmake -f

GOOPTS=CGO_ENABLED=0
BUILDOPTS=-ldflags="-s -w" -a -gcflags=all=-l

all: clean build

build:
	${GOOPTS} go build ${BUILDOPTS} -o seed-fortune      aleesa-phrases-lib.go seed-fortune.go
	${GOOPTS} go build ${BUILDOPTS} -o seed-proverb      aleesa-phrases-lib.go seed-proverb.go
	${GOOPTS} go build ${BUILDOPTS} -o seed-friday       aleesa-phrases-lib.go seed-friday.go
	${GOOPTS} go build ${BUILDOPTS} -o aleesa-phrases-go aleesa-phrases-lib.go aleesa-phrases.go

clean:
	go clean

wipe:
	go clean
	rm -rf go.{mod,sum}

prep:
	go mod init main
	go mod tidy -compat=1.17

vendor:
	rm -rf vendor
	go mod vendor

# vim: set ft=make noet ai ts=4 sw=4 sts=4:
