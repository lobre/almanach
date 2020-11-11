.PHONY: lin win gen clean

lin: gen
	@echo "==> Building App..." && \
	go build

win: gen
	@echo "==> Building App in MinGW container..." && \
	docker run --rm -it -v "$(PWD)":/go/src -v "$(GOPATH)/pkg/mod":/go/pkg/mod lobre/go-mingw build

gen:
	@echo "==> Generating files..." && \
	go generate

clean:
	@echo "==> Cleaning..." && \
	rm -f almanach almanach.exe almanach.db pkged.go

