# lambda-container-exec

For container running on AWS Lambda.
It exec code downloaded from S3.

## Setup

1. Place lambda-container-exec binary at `/main` in your container image.

   ```docker
   COPY lambda-container-exec /main
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

If source code is a tarball, lambda-container-exec will extract it and exec `bootstrap` in it.

## Author

https://github.com/handlename

## License

MIT
