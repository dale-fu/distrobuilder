package generators

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	incusArch "github.com/lxc/incus/v7/shared/osarch"

	"github.com/lxc/distrobuilder/v3/image"
	"github.com/lxc/distrobuilder/v3/shared"
)

type fstab struct {
	common
}

// RunLXC doesn't support the fstab generator.
func (g *fstab) RunLXC(img *image.LXCImage, target shared.DefinitionTargetLXC) error {
	return errors.New("fstab generator not supported for LXC")
}

// RunIncus writes to /etc/fstab.
func (g *fstab) RunIncus(img *image.IncusImage, target shared.DefinitionTargetIncus) error {
	f, err := os.Create(filepath.Join(g.sourceDir, "etc/fstab"))
	if err != nil {
		return fmt.Errorf("Failed to create file %q: %w", filepath.Join(g.sourceDir, "etc/fstab"), err)
	}

	defer f.Close()

	fs := target.VM.Filesystem

	if fs == "" {
		fs = "ext4"
	}

	options := "defaults"

	if fs == "btrfs" {
		options = fmt.Sprintf("%s,subvol=@", options)
	}

	// Determine the boot partition fstab entry based on architecture.
	// s390x uses an ext4 /boot partition; ppc64le uses a raw PReP partition
	// (no filesystem, no fstab entry); all others use a vfat /boot/efi ESP.
	archID, _ := incusArch.ArchitectureID(g.def.Image.Architecture)

	var bootEntry string

	switch archID {
	case incusArch.ARCH_64BIT_S390_BIG_ENDIAN:
		bootEntry = "LABEL=boot    /boot     ext4  defaults  0 2\n"
	case incusArch.ARCH_64BIT_POWERPC_LITTLE_ENDIAN:
		// PReP partition has no filesystem; no fstab entry needed.
	default:
		bootEntry = "LABEL=UEFI    /boot/efi vfat  defaults  0 0\n"
	}

	_, err = fmt.Fprintf(f, "LABEL=rootfs  /         %s  %s  0 0\n%s", fs, options, bootEntry)
	if err != nil {
		return fmt.Errorf("Failed to write string to file %q: %w", filepath.Join(g.sourceDir, "etc/fstab"), err)
	}

	return nil
}

// Run does nothing.
func (g *fstab) Run() error {
	return nil
}
