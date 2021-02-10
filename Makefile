VERSION=$(shell cat ./VERSION)
PROJECT_USERNAME=handlename
PROJECT_REPONAME=lambda-container-exec
DIST_DIR=dist
MAKEFILE_PATH=$(abspath $(dir $(lastword $(MAKEFILE_LIST))))

export GOOS=linux
export GOARCH=amd64

test:
	go test -v ./...

run: .aws-lambda-rie dist
	docker run \
		-v $(MAKEFILE_PATH)/.aws-lambda-rie:/aws-lambda \
		-v $(MAKEFILE_PATH)/dist/lambda-container-exec_v$(VERSION)_linux_amd64:/main \
		-e AWS_REGION=${AWS_REGION} \
		-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
		-e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
		-e AWS_SESSION_TOKEN=${AWS_SESSION_TOKEN} \
		-e CONTAINER_EXEC_LOG_LEVEL=DEBUG \
		-e CONTAINER_EXEC_SRC=$(CONTAINER_EXEC_SRC) \
		--entrypoint /aws-lambda/aws-lambda-rie \
		-p 9000:8080 \
		alpine:latest \
		/main

.PHONY: .aws-lambda-rie
.aws-lambda-rie:
	# https://github.com/aws/aws-lambda-runtime-interface-emulator#test-an-image-without-adding-rie-to-the-image
	mkdir -p $@
	curl -sL https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/download/v1.0/aws-lambda-rie -o $@/aws-lambda-rie
	chmod +x $@/aws-lambda-rie

.PHONY: dist
dist: clean
	mkdir -p $(DIST_DIR)
	go build \
		-ldflags '-X main.version=$(VERSION)' \
		-o '$(DIST_DIR)/$(PROJECT_REPONAME)_v$(VERSION)_$(GOOS)_$(GOARCH)' \
		.

.PHONY: build-docker-image
build-docker-image: dist
	docker build \
	  --rm \
	  --build-arg VERSION=$(VERSION) \
	  --tag $(PROJECT_USERNAME)/$(PROJECT_REPONAME):$(VERSION) \
	  --tag ghcr.io/$(PROJECT_USERNAME)/$(PROJECT_REPONAME):$(VERSION) \
	  .

.PHONY: push-docker-image
push-docker-image:
	docker push ghcr.io/$(PROJECT_USERNAME)/$(PROJECT_REPONAME):$(VERSION)

.PHONY: release
release:
	git tag v$(VERSION)
	git push origin v$(VERSION)

.PHONY: clean
clean:
	rm -rf $(PROJECT_REPONAME) $(DIST_DIR)/*
