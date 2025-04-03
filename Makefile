INPUTS=examples/hello-world.bf examples/fib.bf

all: main

.PHONY: test
test: main $(INPUTS)
	echo; for f in $(INPUTS); do echo "*** $$f"; ./gbf $$f; echo; done

main: main.go
	go build .
