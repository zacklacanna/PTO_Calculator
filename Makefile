APP := pto_calculator

.PHONY: run build test fmt clean

run:
	go run . --type tui

build:
	go build -o $(APP) .

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	rm -f $(APP)
