PLUGIN_ID ?= com.github.yiwkr.mattermost-plugin-janken
PLUGIN_VERSION ?= 0.0.1
BUNDLE_NAME ?= $(PLUGIN_ID)_$(PLUGIN_VERSION).tar.gz
GO ?= $(shell command -v go 2>/dev/null)
DEP ?= $(shell command -v dep 2>/dev/null)
CURL ?= $(shell command -v curl 2>/dev/null)

ASSETS_DIR ?= assets

.PHONY: dep
dep:
	cd server && $(DEP) ensure

.PHONY: build
build: dep
	mkdir -p dist
	cd server && env GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o dist/plugin-linux-amd64
	cd server && env GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-s -w" -o dist/plugin-darwin-amd64
	cd server && env GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-s -w" -o dist/plugin-windows-amd64.exe

.PHONY: bundle
bundle:
	rm -rf dist/
	mkdir -p dist/$(PLUGIN_ID)
	cp plugin.json dist/$(PLUGIN_ID)/

ifneq ($(wildcard $(ASSETS_DIR)/.),)
	cp -r $(ASSETS_DIR) dist/$(PLUGIN_ID)/
endif

	mkdir -p dist/$(PLUGIN_ID)/server/dist
	cp -r server/dist/* dist/$(PLUGIN_ID)/server/dist

	cd dist && tar -cvzf $(BUNDLE_NAME) $(PLUGIN_ID)

	@echo plugin built at: dist/$(BUNDLE_NAME)

.PHONY: dist
dist: build bundle

.PHONY: deploy
deploy: dist
## It uses the API if appropriate environment variables are defined,
## or copying the files directly to a sibling mattermost-server directory.
ifneq ($(and $(MM_SERVICESETTINGS_SITEURL),$(MM_ADMIN_USERNAME),$(MM_ADMIN_PASSWORD),$(CURL)),)
	@echo "Installing plugin via API"
	$(eval TOKEN := $(shell $(CURL) -i -X POST $(MM_SERVICESETTINGS_SITEURL)/api/v4/users/login -d '{"login_id": "$(MM_ADMIN_USERNAME)", "password": "$(MM_ADMIN_PASSWORD)"}' | grep Token | cut -f2 -d' ' | sed 's/\r//' 2> /dev/null))
	@$(CURL) -i -s -H "Authorization: Bearer $(TOKEN)" -X DELETE $(MM_SERVICESETTINGS_SITEURL)/api/v4/plugins/$(PLUGIN_ID)
	@$(CURL) -i -s -H "Authorization: Bearer $(TOKEN)" -X POST $(MM_SERVICESETTINGS_SITEURL)/api/v4/plugins -F "plugin=@dist/$(BUNDLE_NAME)" -F "force=true" && \
		$(CURL) -s -H "Authorization: Bearer $(TOKEN)" -X POST $(MM_SERVICESETTINGS_SITEURL)/api/v4/plugins/$(PLUGIN_ID)/enable && \
		echo "OK." || echo "Sorry, something went wrong."
else
	@echo "No supported deployment method available. Install plugin manually."
endif

.PHONY: test
test: dep
	$(GO) test -race -gcflags=-l -coverprofile=coverage.out `go list ./...` && go tool cover -html=coverage.out -o coverage.html
