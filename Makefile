BINARY := grg
GOOS := darwin
GOARCH := arm64
BUMP ?= patch

.PHONY: all build test clean next-version release

all: build

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY) .

test:
	go test ./...

clean:
	rm -f $(BINARY)

next-version:
	@git fetch origin --tags >/dev/null 2>&1 || true
	@case "$(BUMP)" in \
		major|minor|patch) ;; \
		*) echo 'BUMP は patch / minor / major のいずれかです。' >&2; exit 1 ;; \
	esac; \
	last=$$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n1); \
	if [ -z "$$last" ]; then echo v0.1.0; exit 0; fi; \
	ver=$${last#v}; \
	major=$$(printf '%s\n' "$$ver" | cut -d. -f1); \
	minor=$$(printf '%s\n' "$$ver" | cut -d. -f2); \
	patch=$$(printf '%s\n' "$$ver" | cut -d. -f3); \
	case "$(BUMP)" in \
		major) major=$$((major + 1)); minor=0; patch=0 ;; \
		minor) minor=$$((minor + 1)); patch=0 ;; \
		patch) patch=$$((patch + 1)) ;; \
	esac; \
	printf 'v%s.%s.%s\n' "$$major" "$$minor" "$$patch"

release:
	@set -e; \
	test -z "$$(git status --porcelain)" || { echo '作業ツリーが dirty です。コミットしてから make release してください。' >&2; exit 1; }; \
	if [ -n "$(VERSION)" ]; then \
		VERSION="$(VERSION)"; \
	else \
		VERSION=$$($(MAKE) -s next-version BUMP="$(BUMP)"); \
	fi; \
	echo "リリース: $$VERSION"; \
	if git rev-parse "refs/tags/$$VERSION" >/dev/null 2>&1; then echo "タグ $$VERSION はすでにあります。" >&2; exit 1; fi; \
	$(MAKE) test; \
	$(MAKE) build; \
	git push origin HEAD; \
	git tag "$$VERSION"; \
	git push origin "refs/tags/$$VERSION"; \
	gh release create "$$VERSION" --title "$$VERSION" --generate-notes "$(BINARY)"
