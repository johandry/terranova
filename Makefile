# Copyright The Terranova Authors

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# 	http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
