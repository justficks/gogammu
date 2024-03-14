BIN_FILENAME=nnsms-checker
REMOTE_USER=test
REMOTE_HOST=nn
REMOTE_PATH=/opt/nnsms/checker

build-checker:
	echo "Compiling for Linux..."
	GOOS=linux GOARCH=amd64 go build -o $(BIN_FILENAME) ./cmd/checker

deploy: build-checker
	echo "Copy $(BIN_FILENAME) to $(REMOTE_HOST):$(REMOTE_PATH) ..."
	scp $(BIN_FILENAME) $(REMOTE_HOST):$(REMOTE_PATH)
	rm $(BIN_FILENAME)
