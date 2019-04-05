SHELL	:= /bin/bash

.PHONY: default 
default: install

.PHONY: all 
all: install

.PHONY: install 
install: fmt test
	go install .

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: fmt
fmt:
	go fmt ./...
	go vet ./...
	go list ./... | xargs -n1 golint

# init:
# 	-@[[ -x $${GOPATH}/bin/govendor ]] || go get -u github.com/kardianos/govendor
# 	@govendor init
# 	@$(MAKE) vendor

# vendor:
# 	@govendor list -no-status +missing | xargs -n1 go get -u
# 	@govendor add +external

# vendor-update:
# 	@govendor update +vendor

# clean:
# 	@govendor remove +vendor

# clean-all: clean
# 	@$(RM) -r ./vendor/
