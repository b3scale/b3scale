######################################################################
# @author      : annika (annika@berlin.ccc.de)
# @file        : Makefile
# @created     : Sunday Aug 16, 2020 19:24:54 CEST
######################################################################


test:
	cd lib/config && go test -v


.PHONY: clean test

clean:
	rm -f $(ODIR)/*.o

