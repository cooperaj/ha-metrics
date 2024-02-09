#PREFIX is environment variable, but if it is not set, then set default value
ifeq ($(PREFIX),)
    PREFIX := /usr
endif

BINARY_NAME=ha-metrics

OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
 
.PHONY: build
build:
	mkdir -p out/bin
	GOPROXY=direct GO111MODULE=auto GOARCH=$(ARCH) GOOS=$(OS) go build -o out/bin/$(OS)/$(BINARY_NAME)

.PHONY: build-deb
build-deb:
	dpkg-buildpackage -us -uc -b -a $(ARCH)

.PHONY: clean
clean:
	go clean
	-rm -R out/

.PHONY: clean-deb
clean-deb:	
	dpkg-buildpackage -Tclean

.PHONY: install
install: 
	install -d $(DESTDIR)$(PREFIX)/bin/
	install -c -m 755 out/bin/$(OS)/$(BINARY_NAME) $(DESTDIR)$(PREFIX)/bin/

	mkdir -p $(DESTDIR)/etc/ha-metrics
	install -c -m 600 conf.toml.dist $(DESTDIR)/etc/ha-metrics/conf.toml

.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)
	rm -f /lib/systemd/system/$(BINARY_NAME).service
	rm -Rf /etc/$(BINARY_NAME)
	systemctl daemon-reload