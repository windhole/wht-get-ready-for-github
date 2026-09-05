BINARY := grg
GOOS := darwin
GOARCH := arm64

.PHONY: all build test clean release

all: build

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY) .

test:
	go test ./...

clean:
	rm -f $(BINARY)

release:
	@test -n "$(VERSION)" || { echo 'VERSION を指定してください。例: make release VERSION=v0.1.0' >&2; exit 1; }
	@test -z "$$(git status --porcelain)" || { echo '作業ツリーが dirty です。コミットしてから make release してください。' >&2; exit 1; }
	@if git rev-parse "refs/tags/$(VERSION)" >/dev/null 2>&1; then echo "タグ $(VERSION) はすでにあります。" >&2; exit 1; fi
	$(MAKE) test
	$(MAKE) build
	git push origin HEAD
	git tag "$(VERSION)"
	git push origin "refs/tags/$(VERSION)"
	gh release create "$(VERSION)" --title "$(VERSION)" --generate-notes "$(BINARY)"
