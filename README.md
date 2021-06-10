SIPS
====

[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/DeedleFake/sips)](https://pkg.go.dev/github.com/DeedleFake/sips)
[![Go Report Card](https://goreportcard.com/badge/github.com/DeedleFake/sips)](https://goreportcard.com/report/github.com/DeedleFake/sips)

*Disclaimer: SIPS is still in early development and is not guaranteed to do much of anything. Although it should function for basic usage, expect bugs, and definitely don't use it for anything that has money associated with it.*

SIPS is a Simple IPFS Pinning Service. It does the bare minimum necessary to present a functional [pinning service][pinning-service-api].

Setup
-----

After installation, SIPS will have no users or tokens in its database. To create some, use the `sipsctl` utility that is provided:

```bash
$ sipsctl users add whateverUsernameYouWant
$ sipsctl tokens add --user whateverUsernameYouWant
```

You can then use that token with a pinning service client to add, remove, and list pins.

[pinning-service-api]: https://ipfs.github.io/pinning-services-api-spec/
