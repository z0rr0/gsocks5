TARGET=gsocks5
DOCKER_TAG=z0rr0/gsocks5
TAG=$(shell git tag | sort -V | tail -1)
LDFLAGS=-X main.Tag=$(TAG)

build:
	go build -ldflags "$(LDFLAGS)" -o $(PWD)/$(TARGET)

fmt:
	gofmt -d .

check_fmt:
	@test -z "`gofmt -l .`" || { echo "ERROR: failed gofmt, for more details run - make fmt"; false; }
	@-echo "gofmt successful"

lint: check_fmt
	go vet $(PWD)/...
	golint -set_exit_status $(PWD)/...
	golangci-lint run ./...

test: check_fmt
	go test -race -cover $(PWD)/...

docker:
	docker build --build-arg LDFLAGS="$(LDFLAGS)" -t $(DOCKER_TAG) .

clean:
	rm -f $(PWD)/$(TARGET)
	find ./ -type f -name "*.out" -delete
