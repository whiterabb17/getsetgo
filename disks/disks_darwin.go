package disks

import (
	"os"
)

func disks() ([]string, error) {
	found_drives := []string{}

	for _, drive := range "abcdefgh" {
		f, err := os.Open("/dev/sd" + string(drive))
		if err == nil {
			found_drives = append(found_drives, "/dev/sd"+string(drive))
			f.Close()
		}
	}

	return found_drives, nil
}
