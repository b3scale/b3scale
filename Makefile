######################################################################
# @author      : annika (annika@berlin.ccc.de)
# @file        : Makefile
# @created     : Sunday Aug 16, 2020 19:24:54 CEST
######################################################################

include Makefile.inc

TARGETS:=all build static clean
SUBDIRS:=cmd/b3scaled cmd/b3scalectl cmd/b3scalenoded

$(TARGETS): subdirs

subdirs: $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $@ $(filter $(TARGETS),$(MAKECMDGOALS))

test:
	go test ./pkg/...

.PHONY: subdirs $(TARGETS) $(SUBDIRS) test

