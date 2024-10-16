.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o wechatbot ./main.go

.PHONY: debug
debug:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -gcflags "all=-N -l" -ldflags '-w' -o wechatbot ./main.go

.PHONY: docker
docker:
	docker build . -t wechatbot:latest
