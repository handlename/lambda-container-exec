VERSION=$(shell cat ./VERSION)
PROJECT_USERNAME=handlename
PROJECT_REPONAME=lambda-container-exec
DIST_DIR=dist
MAKEFILE_PATH=$(abspath $(dir $(lastword $(MAKEFILE_LIST))))

run: build-run-image .aws-lambda-rie dist
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
		$(PROJECT_USERNAME)/$(PROJECT_REPONAME) \
		/main

.PHONY: build-image
build-run-image:
	docker build . -t $(PROJECT_USERNAME)/$(PROJECT_REPONAME):run

.PHONY: .aws-lambda-rie
.aws-lambda-rie:
	# https://github.com/aws/aws-lambda-runtime-interface-emulator#test-an-image-without-adding-rie-to-the-image
	mkdir -p $@
	curl -sL https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/download/v1.0/aws-lambda-rie -o $@/aws-lambda-rie
	chmod +x $@/aws-lambda-rie

.PHONY: dist
dist: clean
	mkdir -p $(DIST_DIR)
	gox \
		-ldflags '-X main.version=$(VERSION)' \
		-os='linux' \
		-arch='amd64' \
		-output='$(DIST_DIR)/$(PROJECT_REPONAME)_v$(VERSION)_{{ .OS }}_{{ .Arch }}' \
		.

.PHONY: clean
clean:
	rm -rf $(PROJECT_REPONAME) $(DIST_DIR)/*
