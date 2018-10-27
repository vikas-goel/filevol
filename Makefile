# Installation Directories
SYSCONFDIR ?=$(DESTDIR)/etc/docker
SYSTEMDIR ?=$(DESTDIR)/usr/lib/systemd/system
GOLANG ?= /usr/bin/go
BINARY ?= docker-filevol-plugin
BINDIR ?=$(DESTDIR)/usr/libexec/docker

export GO15VENDOREXPERIMENT=1

all: plugin-build

.PHONY: plugin-build
plugin-build: src/main.go src/driver.go
	$(GOLANG) build -o $(BINARY) ./src

.PHONY: install
install:
	install -D -m 644 etc/docker/filevol-plugin $(SYSCONFDIR)/filevol-plugin
	install -D -m 644 systemd/docker-filevol-plugin.service $(SYSTEMDIR)/docker-filevol-plugin.service
	install -D -m 644 systemd/docker-filevol-plugin.socket $(SYSTEMDIR)/docker-filevol-plugin.socket
	install -D -m 755 $(BINARY) $(BINDIR)/$(BINARY)

.PHONY: clean
clean:
	rm -f $(BINARY)
