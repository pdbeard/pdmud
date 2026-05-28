BIN := bin
CHAT_SRC := widgets/chat
AFFECTS_SRC := widgets/affects

.PHONY: all build clean run

all: build

build:
	@echo "Building widgets..."
	@mkdir -p $(BIN)
	go build -o $(BIN)/chat ./$(CHAT_SRC)
	go build -o $(BIN)/affects ./$(AFFECTS_SRC)
	@echo "Done."

clean:
	rm -f $(BIN)/chat $(BIN)/affects

run: build
	@./scripts/launch.sh
