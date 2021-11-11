This repo is a fork of https://github.com/gokrazy/kernel.

# gokrazy Odroid XU4/HC1/HC2 kernel repository

This repository holds a pre-built Linux kernel image for Exynos5422-based
Odroid boards (Odroid XU4/HC1/HC2), used by the [gokrazy](https://github.com/gokrazy/gokrazy) project.

The files in this repository are picked up automatically by
`gokr-packer`, so you don’t need to interact with this repository
unless you want to update the kernel to a custom version.

## Updating the kernel

First, install docker.

Then, build a new kernel - take about 5 minutes.
```
go run ./cmd/gokr-rebuild-kernel/kernel.go
```
