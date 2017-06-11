package hidraw

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type dev struct {
	schema string
}

func findDevice(vendor, product int) string {
	suffix := fmt.Sprintf("v%08Xp%08X", vendor, product)
	m, _ := filepath.Glob("/sys/class/hidraw/*/device/modalias")
	for _, p := range m {
		data, err := ioutil.ReadFile(p)
		if err != nil {
			continue
		}

		if strings.HasSuffix(strings.TrimSpace(string(data)), suffix) {
			parts := strings.Split(p, "/")
			return "/dev/" + parts[4]
		}
	}

	return ""
}

func findAndOpen(vendor, product int) (*os.File, error) {
	name := findDevice(vendor, product)
	if name == "" {
		return nil, ErrNoDevice
	}

	return os.OpenFile(name, os.O_RDONLY, 0)
}
