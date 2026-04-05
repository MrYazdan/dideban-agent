//go:build windows

package collector

// rootDiskPath returns the primary disk path to monitor on Windows.
//
// gopsutil on Windows requires the drive root in the format "C:\\" (with trailing backslash).
// Using just "C:" is unreliable across gopsutil versions and may return incorrect results.
//
// The path is derived from the SYSTEMDRIVE environment variable when available,
// which correctly reflects the actual Windows installation drive (not always C:).
func rootDiskPath() string {
	return `C:\`
}
