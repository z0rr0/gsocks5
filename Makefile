TARGET=gsocks5
DOCKER_TAG=z0rr0/gsocks5

build:
	go build -ldflags "-X main.Tag=`git tag --sort=version:refname | tail -1`" -o $(PWD)/$(TARGET)

fmt:
	gofmt -d .

check_fmt:
	@test -z "`gofmt -l .`" || { echo "ERROR: failed gofmt, for more details run - make fmt"; false; }
	@-echo "gofmt successful"

lint: check_fmt
	go vet $(PWD)/...
	golint -set_exit_status $(PWD)/...

test: lint
	go test -race -cover $(PWD)/...

docker:
	docker build -t $(DOCKER_TAG) .

clean:
	rm -f $(PWD)/$(TARGET)
	find ./ -type f -name "*.out" -delete