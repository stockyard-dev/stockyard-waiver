build:
	CGO_ENABLED=0 go build -o waiver ./cmd/waiver/

run: build
	./waiver

test:
	go test ./...

clean:
	rm -f waiver

.PHONY: build run test clean
