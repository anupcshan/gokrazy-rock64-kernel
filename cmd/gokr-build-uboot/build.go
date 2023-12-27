package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

const ubootRev = "0a0ceea2269b983e736b80104f03cc800d1a5e2a"
const ubootTS = 1703441575

const (
	uBootRepo = "https://github.com/u-boot/u-boot"
)

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
	defconfig := exec.Command("make", "ARCH=arm", "odroid-xu3_defconfig")
	defconfig.Stdout = os.Stdout
	defconfig.Stderr = os.Stderr
	if err := defconfig.Run(); err != nil {
		return fmt.Errorf("make defconfig: %v", err)
	}

	f, err := os.OpenFile(".config", os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	// u-boot began failing boot around commit 13819f07ea6c60e87b708755a53954b8c0c99a32.
	// CONFIG_BOARD_LATE_INIT tries to load CROS_EC, which clearly doesn't exist on HC2.
	if _, err := f.Write([]byte("CONFIG_CMD_SETEXPR=y\nCMD_SETEXPR_FMT=y\nCONFIG_BOARD_LATE_INIT=n\n")); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	make := exec.Command("make", "u-boot.bin", "-j"+strconv.Itoa(runtime.NumCPU()))
	make.Env = append(os.Environ(),
		"ARCH=arm",
		"CROSS_COMPILE=arm-linux-gnueabihf-",
		"SOURCE_DATE_EPOCH="+strconv.Itoa(ubootTS),
	)
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return fmt.Errorf("make: %v", err)
	}

	return nil
}

func generateBootScr(bootCmdPath string) error {
	mkimage := exec.Command("./tools/mkimage", "-A", "arm", "-O", "linux", "-T", "script", "-C", "none", "-a", "0", "-e", "0", "-n", "Gokrazy Boot Script", "-d", bootCmdPath, "boot.scr")
	mkimage.Env = append(os.Environ(),
		"ARCH=arm",
		"CROSS_COMPILE=arm-linux-gnueabihf-",
		"SOURCE_DATE_EPOCH=1600000000",
	)
	mkimage.Stdout = os.Stdout
	mkimage.Stderr = os.Stderr
	if err := mkimage.Run(); err != nil {
		return fmt.Errorf("mkimage: %v", err)
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
	tmpDir, err := os.MkdirTemp(os.TempDir(), "u-boot")
	if err != nil {
		log.Fatal(err)
	}

	var bootCmdPath string
	if p, err := filepath.Abs("boot.cmd"); err != nil {
		log.Fatal(err)
	} else {
		bootCmdPath = p
	}

	if err := os.Chdir(tmpDir); err != nil {
		log.Fatal(err)
	}

	for _, cmd := range [][]string{
		{"git", "init"},
		{"git", "remote", "add", "origin", uBootRepo},
		{"git", "fetch", "--depth=1", "origin", ubootRev},
		{"git", "checkout", "FETCH_HEAD"},
	} {
		log.Printf("Running %s", cmd)
		cmdObj := exec.Command(cmd[0], cmd[1:]...)
		cmdObj.Stdout = os.Stdout
		cmdObj.Stderr = os.Stderr
		cmdObj.Dir = tmpDir
		if err := cmdObj.Run(); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("applying patches")
	if err := applyPatches(tmpDir); err != nil {
		log.Fatal(err)
	}

	log.Printf("compiling uboot")
	if err := compile(); err != nil {
		log.Fatal(err)
	}

	log.Printf("generating boot.scr")
	if err := generateBootScr(bootCmdPath); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/u-boot.bin", "u-boot.bin"); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/boot.scr", "boot.scr"); err != nil {
		log.Fatal(err)
	}
}
