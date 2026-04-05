//go:build !windows

package collector

// rootDiskPath returns the primary disk path to monitor on Unix-like systems.
//
// "/" is the root filesystem mount point on Linux, macOS, and BSD variants.
func rootDiskPath() string {
	return "/"
}
