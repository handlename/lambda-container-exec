# lambda-container-exec

For container running on AWS Lambda.
It exec code downloaded from S3.

## Setup

1. Place lambda-container-exec binary at `/main` in your container image.

   ```docker
   FROM ghcr.io/handlename/lambda-container-exec:0.1.0 AS lambda-container-exec
   FROM ...
   COPY --from lambda-container-exec /usr/local/bin/lambda-container-exec /main
   ```
1. Set ENTRYPOINT as `"/main"`

   ```docker
   ENTRYPOINT ["/main"]
   ```
1. Set source code path to environment variable `CONTAINER_EXEC_SRC`

## Source code structure

lambda-container-exec exec `bootstarp` in downloaded source code directory.

```
CONTAINER_EXEC_SRC/
|
`-- bootstrap
`-- ...
```

## Author

https://github.com/handlename

## License

MIT
