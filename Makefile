NAME=aerie
REPO=registry.digitalocean.com/zushco

lint:
	golangci-lint run ./...

test:
	go test -v ./...

tidy:
	go mod tidy

run: tidy
	SERVICE_ACCOUNT_EMAIL=mike@zush.co go run cmd/aerie/main.go

build:
	GOOS=linux GOARCH=amd64 go build -o bin/main cmd/$(NAME)/main.go

image: build
	docker build -t $(NAME) .
	docker tag $(NAME) $(REPO)/$(NAME):latest

deploy: image
	docker buildx build --push \
		--platform linux/amd64 \
		--tag $(REPO)/$(NAME):latest  .
