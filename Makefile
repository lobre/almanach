.PHONY: lin win gen

lin: gen
	@echo "==> Building App..." && \
	go build

win: gen
	@echo "==> Building App in MinGW container..." && \
	docker run --rm -it -v "$(PWD)":/go/src -v "$(GOPATH)/pkg/mod":/go/pkg/mod lobre/go-mingw build

gen:
	@echo "==> Generating files..." && \
	go generate
