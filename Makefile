.PHONY: gifs

all: gifs

TAPES=$(shell ls doc/vhs/*tape)

gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

