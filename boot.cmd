# Hardcoded addresses to load vmlinuz and dtb copied from
# https://www.hardkernel.com/blog-2/upstream-u-boot-for-odroid-xu4/

# Expect all files to be on the first partition of the SD card (2:1)

# Load compressed kernel image
fatload mmc 2:1 0x40008000 vmlinuz

# Emulate cmdline.txt behavior from Raspberry Pi devices.
# Load cmdline.txt into memory (exact location doesn't matter, it shouldn't conflict with any other loads).
fatload mmc 2:1 0x42000000 cmdline.txt
# ... and set string value of var bootargs to it.
# Requires CONFIG_CMD_SETEXPR=y while building u-boot.
setexpr.s bootargs *0x42000000

# Load dtb
fatload mmc 2:1 0x44000000 exynos5422-odroidhc1.dtb
# ... and set fdt addr to it.
fdt addr 0x44000000

# Boot with compressed kernel without initrd
bootz 0x40008000 - 0x44000000
