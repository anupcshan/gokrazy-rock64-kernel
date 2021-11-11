#!/bin/bash

set -eu

mkdir -p ~/gokrazy
mkdir ~/gokrazy/bakery || { echo 'bakery already exists' >&2; exit 1; }
cat > ~/gokrazy/bakery/config.json <<EOT
{
    "Hostname": "gokr-boot-will-inject-the-hostname",
    "Update": {
        "HTTPPassword": "${GOKRAZY_BAKERY_PASSWORD}"
    },
    "DeviceType": "odroidhc1",
    "Packages": [
        "github.com/gokrazy/breakglass",
        "github.com/gokrazy/bakery/cmd/bake",
        "github.com/gokrazy/timestamps",
        "github.com/gokrazy/serial-busybox"
    ],
    "PackageConfig": {
        "github.com/gokrazy/breakglass": {
            "CommandLineFlags": [
                "-authorized_keys=/etc/breakglass.authorized_keys"
            ],
            "ExtraFileContents": {
                "/etc/breakglass.authorized_keys": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPrXgBg9kOZuG7j8ZkguxXbsJ5/bC1oILizs/BPsrF2c anupc@devbox"
            }
        }
    },
    "SerialConsole": "disabled",
    "KernelPackage": "github.com/anupcshan/gokrazy-rock64-kernel",
    "FirmwarePackage": "",
    "EEPROMPackage": ""
}
EOT
