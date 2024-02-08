#PREFIX is environment variable, but if it is not set, then set default value
ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

BINARY_NAME=ha-metrics

OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
 
.PHONY: build
build:
	mkdir -p out/bin
	GOARCH=$(ARCH) GOOS=$(OS) go build -o out/bin/$(OS)/$(BINARY_NAME)

.PHONY: clean
clean:
	go clean
	rm -R out/

.PHONY: install
install: 
	install -d $(DESTDIR)$(PREFIX)/bin/
	install -c -m 755 out/bin/$(OS)/$(BINARY_NAME) $(DESTDIR)$(PREFIX)/bin/

	install -c -m 644 service/systemd/service.service /etc/systemd/system/$(BINARY_NAME).service
    install -d /etc/systemd/system/$(BINARY_NAME).service.d
    :install -c -m 600 service/systemd/service.conf /etc/systemd/system/$(BINARY_NAME).service.d/override.conf
	systemctl daemon-reload

.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)
	rm -f /etc/systemd/system/$(BINARY_NAME).service
	rm -Rf /etc/systemd/system/$(BINARY_NAME).servce.d
	systemctl daemon-reload