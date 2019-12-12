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

.PHONY: install 
install: fmt test
	go install .

.PHONY: test
test:
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic -v ./...

.PHONY: fmt
fmt:
	go fmt ./...
	go vet ./...

.PHONY: check-fmt
check-fmt:
	@files=$$(GO111MODULE=off go fmt ./...); \
	if [[ -n $${files} ]]; then echo "Go fmt found errors in the following files:\n$${files}\n"; exit 1; fi
	@go vet ./...