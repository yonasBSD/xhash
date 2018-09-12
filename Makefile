SSL_LIB_DIR = $(shell ls -d /usr/local/ssl/lib64)
SSL_LIB_DIR ?= /usr/local/ssl/lib

.PHONY: all

all:
	go build -ldflags "-L /usr/local/ssl/lib -extldflags -Wl,-rpath=${SSL_LIB_DIR}"

install: xhash
	@install -m 0755 xhash /usr/local/bin
