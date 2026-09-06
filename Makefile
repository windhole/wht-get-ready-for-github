BINARY := dist/grg
GOOS := darwin
GOARCH := arm64
PKG := ./src

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: all build test clean show-version release release-minor release-major computed-version do-release

all: build

build:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

test:
	go test $(PKG)/...

clean:
	rm -rf dist

show-version:
	@git fetch origin --tags >/dev/null 2>&1 || true
	@last=$$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n1); \
	if [ -z "$$last" ]; then \
		echo '現在: （タグなし）'; \
		echo '次の patch: v0.1.0'; \
		echo '次の minor: v0.1.0'; \
		echo '次の major: v0.1.0'; \
		exit 0; \
	fi; \
	ver=$${last#v}; \
	major=$$(printf '%s\n' "$$ver" | cut -d. -f1); \
	minor=$$(printf '%s\n' "$$ver" | cut -d. -f2); \
	patch=$$(printf '%s\n' "$$ver" | cut -d. -f3); \
	echo "現在: $$last"; \
	echo "次の patch: v$$major.$$minor.$$((patch + 1))"; \
	echo "次の minor: v$$major.$$((minor + 1)).0"; \
	echo "次の major: v$$((major + 1)).0.0"

computed-version:
	@git fetch origin --tags >/dev/null 2>&1 || true
	@last=$$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n1); \
	kind="$(KIND)"; \
	if [ -z "$$last" ]; then echo v0.1.0; exit 0; fi; \
	ver=$${last#v}; \
	major=$$(printf '%s\n' "$$ver" | cut -d. -f1); \
	minor=$$(printf '%s\n' "$$ver" | cut -d. -f2); \
	patch=$$(printf '%s\n' "$$ver" | cut -d. -f3); \
	case "$$kind" in \
		major) printf 'v%s.0.0\n' $$((major + 1)) ;; \
		minor) printf 'v%s.%s.0\n' "$$major" $$((minor + 1)) ;; \
		patch) printf 'v%s.%s.%s\n' "$$major" "$$minor" $$((patch + 1)) ;; \
		*) echo '内部エラー: KIND が不正です。' >&2; exit 1 ;; \
	esac

release:
	@$(MAKE) do-release KIND=patch

release-minor:
	@$(MAKE) do-release KIND=minor

release-major:
	@$(MAKE) do-release KIND=major

do-release:
	@set -e; \
	test -z "$$(git status --porcelain)" || { echo '作業ツリーが dirty です。コミットしてから make release してください。' >&2; exit 1; }; \
	VERSION=$$($(MAKE) -s computed-version KIND="$(KIND)"); \
	echo "リリース: $$VERSION"; \
	if git rev-parse "refs/tags/$$VERSION" >/dev/null 2>&1; then echo "タグ $$VERSION はすでにあります。" >&2; exit 1; fi; \
	$(MAKE) test; \
	$(MAKE) build VERSION=$$VERSION; \
	git push origin HEAD; \
	git tag "$$VERSION"; \
	git push origin "refs/tags/$$VERSION"; \
	gh release create "$$VERSION" --title "$$VERSION" --generate-notes "$(BINARY)"
