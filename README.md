# contextual

[![build status](https://img.shields.io/circleci/project/nbio/contextual/master.svg)](https://circleci.com/gh/nbio/contextual)
[![godoc](http://img.shields.io/badge/docs-GoDoc-blue.svg)](https://godoc.org/github.com/nbio/contextual)

Supercharge [Go](https://golang./org) `Dialer` with `DialContext`, providing support for `context.Context` in libraries with only `Dial` or `DialTimeout` methods. Extracted from and used in production at [domainr.com](https://domainr.com).

## Install

`go get github.com/nbio/contextual`

## Usage

```go
// Example context
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

// SSH into a proxy host
client, err := ssh.Dial("tcp", "yourserver.com:22", sshConfig)
if err != nil {
    return err
}

// Wrap the ssh.Client in contextual.Dialer to get DialContext
conn, err := contextual.Dialer{client}.DialContext(ctx, "tcp", "otherserver.com:12345")
if err != nil {
    return err // Could be context.DeadlineExceeded or context.Canceled
}

// Do stuff
```

## Keywords

go, golang, context

## Author

© 2017 nb.io, LLC
