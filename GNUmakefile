BIN	= xhash

# Optionally create these hard-links
ALL	= b2sum b3sum md5sum sha1sum sha256sum sha512sum

GO	:= go
CGO_ENABLED := 0

.PHONY: all
all:	$(BIN)

$(BIN): *.go
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build

.PHONY: test
test:
	$(GO) vet
	$(GO) test -v

.PHONY: bench
bench:
	$(GO) test -bench=.

.PHONY: clean
clean:
	$(GO) clean
	rm -f $(ALL)

.PHONY: gen
gen:
	rm -f go.mod go.sum
	$(GO) mod init $(BIN)
	$(GO) mod tidy

euid	= $(shell id -u)
ifeq ($(euid),0)
BINDIR	= /usr/local/bin
else
BINDIR	= $(HOME)/bin
endif

.PHONY: install
install: $(BIN)
	@mkdir -p $(BINDIR)
	install -s -m 0755 $(BIN) $(BINDIR)

.PHONY: uninstall
uninstall:
	rm -f $(BINDIR)/$(BIN)

$(ALL):	$(BIN)
	@for f in $(ALL) ; do ln -f $(BIN) $$f ; done

.PHONY: ALL
ALL:	$(ALL)

.PHONY: install-all
install-all: $(ALL)
	@for f in $(ALL) ; do install -m 0755 $$f $(BINDIR) ; done

.PHONY: uninstall-all
uninstall-all:
	@for f in $(ALL) ; do rm -f $(BINDIR)/$$f ; done
