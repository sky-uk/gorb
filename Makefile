.PHONY: all binary container push clean format

files := $(shell find . -path ./vendor -prune -o -name '*.go' -print)

all: push

# 0.0 shouldn't clobber any release builds
TAG = 0.0
PREFIX = kobolog/gorb

binary:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w' -o docker/gorb

container: binary
	docker build -t $(PREFIX):$(TAG) docker

push: container
	docker push $(PREFIX):$(TAG)

clean:
	rm -f docker/gorb

format:
	goimports -w $(files)
