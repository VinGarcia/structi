
GOBIN=$(shell go env GOPATH)/bin

test: setup
	$(GOBIN)/richgo test $(args) ./...

lint: setup
	go vet -structtag=false ./...
	$(GOBIN)/staticcheck ./...

# Update adapters to use a new ksql tag
version=
update:
	git tag $(version)
	git push origin $(version)

setup: $(GOBIN)/richgo $(GOBIN)/staticcheck
$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
$(GOBIN)/richgo:
	go install github.com/kyoh86/richgo@latest

# # # # # Examples # # # # #

from_env_vars:
	go run ./examples/from_env_vars/.

from_map:
	go run ./examples/from_map/.

get_struct_info:
	go run ./examples/get_struct_info/.

get_nested_struct_info:
	go run ./examples/get_nested_struct_info/.
