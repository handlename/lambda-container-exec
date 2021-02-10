.PHONY: upload
upload:
	tar czvf ../code.tgz .
	aws s3 cp ../code.tgz $(CONTAINER_EXEC_SRC)
