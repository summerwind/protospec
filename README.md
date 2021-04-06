# protospec

*protospec* is an extensible conformance testing tool for internet protocols. The tests are called '*spec*' and can be written declaratively in YAML format and shared as OCI Artifacts.

*This project is currently in the 'pathfinder' phase. Please be aware that incompatible changes may be made without notice.*

## Install

Please download the latest binary from the [release page](https://github.com/summerwind/protospec/releases) and place it in the path of your choice.

The container image can be downloaded as follows.

```
$ docker pull ghcr.io/summerwind/protospec:latest
```

## Usage

To run protospec, you first need to download test files called '*spec bundle*'. You can then run the tests with the spec bundle.

Here is an example of how to download and run the tests of [RFC7540](https://tools.ietf.org/html/rfc7540).

### Downloading the spec bundle

The spec bundle can be downloaded using the `protospec pull` command as follows. This command will download it to the `spec` directory under the current directory.

```
$ protospec pull ghcr.io/summerwind/protospec-bundle-http2-rfc7540
```

### Running tests using the spec bundle

You can now run the tests using the spec bundle you downloaded. In this example, you will run the test against an HTTP/2 server running on port 8080 of localhost.

```
$ protospec run -p 8080 ./spec
```

When you run the test, you will get the following results.

```
3.5: HTTP/2 Connection Preface
#1: Sends client connection preface
#2: Sends invalid connection preface

4.1: Frame Format
#1: Sends a frame with unknown type
#2: Sends a frame with undefined flag
#3: Sends a frame with reserved field bit

...

8.2: Server Push
#1: Sends a PUSH_PROMISE frame

94 tests, 93 passed, 0 skipped, 1 failed
```

You can use the `-t` option to run only specific tests.

```
$ protospec run -p 8080 -t 3.5/1 ./spec
```

## License

protospec is made available under MIT license.
