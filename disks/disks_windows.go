package disks

import "os"

func disks() ([]string, error) {
	found_drives := []string{}

	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		f, err := os.Open(string(drive) + ":\\")
		if err == nil {
			found_drives = append(found_drives, string(drive)+":\\")
			f.Close()
		}
	}
	return found_drives, nil
}
