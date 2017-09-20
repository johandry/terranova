.PHONY: default test vet golint build init vendor

default: install

all: vendor install

install: vet golint test
	@go install .

test:
	@go list ./... | grep -v -E '^github.com/johandry/platformer/vendor' | xargs -n1 go test -cover

vet:
	@go list ./... | grep -v -E '^github.com/johandry/platformer/vendor' | xargs -n1 go vet -v

golint:
	-@[[ -x $${GOPATH}/bin/golint ]] || go get github.com/golang/lint/golint
	@go list ./... | grep -v -E '^github.com/johandry/platformer/vendor' | xargs -n1 golint

init:
	-@[[ -x $${GOPATH}/bin/govendor ]] || go get -u github.com/kardianos/govendor
	@govendor init
	@$(MAKE) vendor

vendor:
	@govendor list -no-status +missing | xargs -n1 go get -u
	@govendor add +external

vendor-update:
	@govendor update +vendor

clean:
	@govendor remove +vendor

clean-all: clean
	@$(RM) -r ./vendor/
