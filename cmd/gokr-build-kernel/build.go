package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	_ "embed"
)

//go:embed config.txt
var configContents []byte

// see https://www.kernel.org/releases.json
var latest = "https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.7.8.tar.xz"

const firmwareSource = "https://git.kernel.org/pub/scm/linux/kernel/git/firmware/linux-firmware.git/plain/%s?id=%s"
const firmwareRevision = "eb8ea1b46893c42edbd516f971a93b4d097730ab"
const firmwareLocation = "/tmp/firmware"

var firmwareFiles = []string{}

func downloadKernel() error {
	out, err := os.Create(filepath.Base(latest))
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(latest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		return fmt.Errorf("unexpected HTTP status code for %s: got %d, want %d", latest, got, want)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return out.Close()
}

func downloadFirmware() error {
	for _, firmwareFile := range firmwareFiles {
		dir := filepath.Dir(firmwareFile)
		if dir != "" {
			if err := os.MkdirAll(filepath.Join(firmwareLocation, dir), 0755); err != nil {
				return err
			}
		}
		out, err := os.Create(filepath.Join(firmwareLocation, firmwareFile))
		if err != nil {
			return err
		}
		defer out.Close()
		resp, err := http.Get(fmt.Sprintf(firmwareSource, firmwareFile, firmwareRevision))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if got, want := resp.StatusCode, http.StatusOK; got != want {
			return fmt.Errorf("unexpected HTTP status code for %s: got %d, want %d", firmwareFile, got, want)
		}
		if _, err := io.Copy(out, resp.Body); err != nil {
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
	}

	return nil
}

func applyPatches(srcdir string) error {
	patches, err := filepath.Glob("*.patch")
	if err != nil {
		return err
	}
	for _, patch := range patches {
		log.Printf("applying patch %q", patch)
		f, err := os.Open(patch)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd := exec.Command("patch", "-p1")
		cmd.Dir = srcdir
		cmd.Stdin = f
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		f.Close()
	}

	return nil
}

func compile() error {
	f, err := os.OpenFile(".config", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(configContents); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	olddefconfig := exec.Command("make", "ARCH=arm64", "oldconfig")
	olddefconfig.Stdout = os.Stdout
	olddefconfig.Stderr = os.Stderr
	if err := olddefconfig.Run(); err != nil {
		return fmt.Errorf("make olddefconfig: %v", err)
	}

	make := exec.Command("make", "Image", "dtbs", "-j"+strconv.Itoa(runtime.NumCPU()))
	make.Env = append(os.Environ(),
		"ARCH=arm64",
		"CROSS_COMPILE=aarch64-linux-gnu-",
		"KBUILD_BUILD_USER=gokrazy",
		"KBUILD_BUILD_HOST=docker",
		"KBUILD_BUILD_TIMESTAMP=Wed Mar  1 20:57:29 UTC 2017",
	)
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return fmt.Errorf("make: %v", err)
	}

	return nil
}

func copyFile(dest, src string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	st, err := in.Stat()
	if err != nil {
		return err
	}
	if err := out.Chmod(st.Mode()); err != nil {
		return err
	}
	return out.Close()
}

func main() {
	log.Printf("downloading firmware")
	if err := os.MkdirAll(firmwareLocation, 0755); err != nil {
		log.Fatal(err)
	}

	if err := downloadFirmware(); err != nil {
		log.Fatal(err)
	}

	log.Printf("downloading kernel source: %s", latest)
	if err := downloadKernel(); err != nil {
		log.Fatal(err)
	}

	log.Printf("unpacking kernel source")
	untar := exec.Command("tar", "xf", filepath.Base(latest))
	untar.Stdout = os.Stdout
	untar.Stderr = os.Stderr
	if err := untar.Run(); err != nil {
		log.Fatalf("untar: %v", err)
	}

	srcdir := strings.TrimSuffix(filepath.Base(latest), ".tar.xz")

	log.Printf("applying patches")
	if err := applyPatches(srcdir); err != nil {
		log.Fatal(err)
	}

	if err := os.Chdir(srcdir); err != nil {
		log.Fatal(err)
	}

	log.Printf("compiling kernel")
	if err := compile(); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/vmlinuz", "arch/arm64/boot/Image"); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/rk3328-rock64.dtb", "arch/arm64/boot/dts/rockchip/rk3328-rock64.dtb"); err != nil {
		log.Fatal(err)
	}
}
