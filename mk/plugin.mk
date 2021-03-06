# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * SRCS       - plugin source files
# * PLUGIN     - plugin binary file
# * GOPKG      - name of the go package
# * DEBUG      - set this to true to compile with debug symbols
# * CLEANFILES - any additional file to remove on clean
# targets:
# * all   - build $(PLUGIN) from $(SRCS) [DEFAULT]
# * clean - remove $(PLUGIN)
# * test  - run $(GOPKG) test
# * lint  - run source code linter

.DEFAULT_GOAL := all

ifndef PLUGIN
  $(error PLUGIN must be set when including plugin.mk)
endif

ifdef DEBUG
  DFLAGS := -gcflags='all=-N -l'
else
  DFLAGS :=
endif

$(PLUGIN): $(SRCS) ; go build $(DFLAGS) -o $(PLUGIN)

.PHONY: all
all: $(PLUGIN)

.PHONY: clean
clean: ; $(RM) $(PLUGIN) $(CLEANFILES)

.PHONY: test
test: ; go test -v -cover -race $(GOPKG)

.PHONY: lint
lint: ; golangci-lint run
