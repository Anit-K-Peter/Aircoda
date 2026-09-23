PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
BINARY = radio

.PHONY: all build install uninstall test clean

all: build

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/radio

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 0755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)

test:
	go test -v ./...

clean:
	rm -f $(BINARY)
