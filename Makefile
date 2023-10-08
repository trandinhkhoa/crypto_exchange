build:
	go build -gcflags "all=-N -l" -o bin/exchange
# go build -o bin/exchange

# run depends on build = before run will run build then execute ./bin/exchange
run: build
		./bin/exchange $(ARGS)

# -v verbose output
# ./... (go command line) : refer to all packages and sub-packages within the current dir
test:
		go test -v ./...

