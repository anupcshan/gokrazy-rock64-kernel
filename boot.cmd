echo "Loading kernel..."

# Load compressed kernel image
load ${devtype} ${devnum}:${bootpart} ${kernel_addr_r} vmlinuz

# Emulate cmdline.txt behavior from Raspberry Pi devices.
# Load cmdline.txt into memory (exact location doesn't matter, it shouldn't conflict with any other loads).
load ${devtype} ${devnum}:${bootpart} 0x42000000 cmdline.txt
# ... and set string value of var bootargs to it.
# Requires CONFIG_CMD_SETEXPR=y while building u-boot.
setexpr.s bootargs *0x42000000

# Load dtb
setenv fdtfile rk3328-rock64.dtb
load ${devtype} ${devnum}:${bootpart} ${fdt_addr_r} ${fdtfile}
# ... and set fdt addr to it.
fdt addr ${fdt_addr_r}

# Boot with compressed kernel without initrd
booti ${kernel_addr_r} - ${fdt_addr_r}
