echo "Loading kernel ..."

# Load compressed kernel image
load ${devtype} ${devnum}:${bootpart} ${kernel_addr_r} vmlinuz

# Emulate cmdline.txt behavior from Raspberry Pi devices.
# Load cmdline.txt into memory (exact location doesn't matter, it shouldn't conflict with any other loads).
load ${devtype} ${devnum}:${bootpart} ${ramdisk_addr_r} cmdline.txt
setexpr cmdline_end ${ramdisk_addr_r} + ${filesize}
# Write 0 byte to the end of cmdline.txt (to terminate the string).
mw.w ${cmdline_end} 0 1
# ... and set string value of var bootargs to it.
# Requires CONFIG_CMD_SETEXPR=y while building u-boot.
setexpr.s bootargs *${ramdisk_addr_r}

echo "Boot args: ${bootargs}"

# Load dtb
setenv fdtfile rk3328-rock64.dtb
load ${devtype} ${devnum}:${bootpart} ${fdt_addr_r} ${fdtfile}
# ... and set fdt addr to it.
fdt addr ${fdt_addr_r}

echo "Booting kernel ..."

# Boot with compressed kernel without initrd
booti ${kernel_addr_r} - ${fdt_addr_r}
