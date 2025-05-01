MAIN := ./cmd/web
TARGET := bin/snippetbox.mabona3.net

always: build

$(TARGET): build
build:
	@go build -o $(TARGET) $(MAIN)

run: $(TARGET)
	@./$(TARGET)

test:
	@go test ./...

clean:
	@rm bin/*
