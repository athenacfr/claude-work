BINARY      := cw
UNAME_S     := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
INSTALL_DIR ?= /usr/local/bin
else
INSTALL_DIR ?= $(HOME)/.local/bin
endif
INSTALL     := $(INSTALL_DIR)/$(BINARY)

.PHONY: build install clean

build:
	go build -o bin/$(BINARY) .

install: build
	rm -f $(INSTALL)
	cp bin/$(BINARY) $(INSTALL)

clean:
	rm -f bin/$(BINARY)
