BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BASEPKG = github.com/emerishq/balcheck

balcheck:
	go build -o build/balcheck -ldflags='-X main.Version=${BRANCH}-${COMMIT}' ${EXTRAFLAGS} ${BASEPKG}/cmd/balcheck