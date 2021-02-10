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

See also AWS official documentation.
https://docs.aws.amazon.com/ja_jp/lambda/latest/dg/go-image.html

## Source code structure

lambda-container-exec downloads source code from `$CONTAINER_EXEC_SRC` as S3 path,
and exec `bootstarp` in extracted source code.

Please check ./exmaple/code directory.

## Author

https://github.com/handlename

## License

MIT
