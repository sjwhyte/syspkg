// Package apt provides a package manager implementation for Debian-based systems using
// Advanced Package Tool (APT) as the underlying package management tool.
package apt

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"

	"github.com/sjwhyte/syspkg/manager"
)

// ParseInstallOutput parses the output of `apt install packageName` command and returns a list of installed packages.
// It extracts the package name, package architecture, and version from the lines that start with "Setting up ".
// Example msg:
//
//	Preparing to unpack .../openssl_3.0.2-0ubuntu1.9_amd64.deb ...
//	Unpacking openssl (3.0.2-0ubuntu1.9) over (3.0.2-0ubuntu1.8) ...
//	Setting up libssl3:amd64 (3.0.2-0ubuntu1.9) ...
//	Setting up libssl3:i386 (3.0.2-0ubuntu1.9) ...
//	Setting up libssl-dev:amd64 (3.0.2-0ubuntu1.9) ...
//	Setting up openssl (3.0.2-0ubuntu1.9) ...
//	Processing triggers for man-db (2.10.2-1) ...
//	Processing triggers for libc-bin (2.35-0ubuntu3.1) ...
func ParseInstallOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	packageInfoPattern := regexp.MustCompile(`Setting up ([\w\d.-]+):?([\w\d]+)? \(([\w\d\.-]+)\)`)

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("apt: %s", line)
		}

		match := packageInfoPattern.FindStringSubmatch(line)

		if len(match) == 4 {
			name := match[1]
			arch := strings.TrimPrefix(match[2], ":") // Remove the colon prefix from the architecture
			version := match[3]

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Arch:           arch,
				Version:        version,
				NewVersion:     version,
				Status:         manager.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParseDeletedOutput parses the output of `apt remove packageName` command
// and returns a list of removed packages.
func ParseDeletedOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("apt: %s", line)
		}

		// TODO: rewrite this using regexp
		if strings.HasPrefix(line, "Removing") {
			parts := strings.Fields(line)
			if opts.Verbose {
				log.Printf("apt: parts: %s", parts)
			}
			var name, arch string
			if strings.Contains(parts[1], ":") {
				name = strings.Split(parts[1], ":")[0]
				arch = strings.Split(parts[1], ":")[1]
			} else {
				name = parts[1]
			}

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        strings.Trim(parts[2], "()"),
				NewVersion:     "",
				Category:       "",
				Arch:           arch,
				Status:         manager.PackageStatusAvailable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParseFindOutput parses the output of `apt search packageName` command
// and returns a list of available packages that match the search query. It extracts package
// information such as name, version, architecture, and category from the
// output, and stores them in a list of manager.PackageInfo objects.
//
// The output format is expected to be similar to the following example:
//
// Sorting...
// Full Text Search...
// zutty/jammy 0.11.2.20220109.192032+dfsg1-1 amd64
// Efficient full-featured X11 terminal emulator
// zvbi/jammy 0.2.35-19 amd64
// Vertical Blanking Interval (VBI) utilities
//
// The function first removes the "Sorting..." and "Full Text Search..."
// lines, and then processes each package entry line to extract relevant
// information.
func ParseFindOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo
	var packagesDict = make(map[string]manager.PackageInfo)

	msg = strings.TrimPrefix(msg, "Sorting...\nFull Text Search...\n")

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")

	// split output by empty lines
	var lines []string = strings.Split(msg, "\n\n")

	for _, line := range lines {
		if regexp.MustCompile(`^[\w\d-]+/[\w\d-,]+`).MatchString(line) {
			parts := strings.Fields(line)

			// if name is empty, it might be not what we want
			if parts[0] == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           strings.Split(parts[0], "/")[0],
				Version:        parts[1],
				NewVersion:     parts[1],
				Category:       strings.Split(parts[0], "/")[1],
				Arch:           parts[2],
				PackageManager: pm,
			}

			packagesDict[packageInfo.Name] = packageInfo
		}
	}

	if len(packagesDict) == 0 {
		return packages
	}

	packages, err := getPackageStatus(packagesDict)
	if err != nil {
		log.Printf("apt: getPackageStatus error: %s\n", err)
	}

	return packages
}

// ParseListInstalledOutput parses the output of `dpkg-query -W -f '${binary:Package} ${Version}\n'` command
// and returns a list of installed packages. It extracts the package name, version,
// and architecture from the output and stores them in a list of manager.PackageInfo objects.
func ParseListInstalledOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(string(msg), "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)

			// if name is empty, it might be not what we want
			if parts[0] == "" {
				continue
			}
			var name, arch string
			if strings.Contains(parts[0], ":") {
				name = strings.Split(parts[0], ":")[0]
				arch = strings.Split(parts[0], ":")[1]
			} else {
				name = parts[0]
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        parts[1],
				Status:         manager.PackageStatusInstalled,
				Arch:           arch,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParseListUpgradableOutput parses the output of `apt list --upgradable` command
// and returns a list of upgradable packages. It extracts the package name, version, new version,
// category, and architecture from the output and stores them in a list of manager.PackageInfo objects.
func ParseListUpgradableOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// Listing...
	// cloudflared/unknown 2023.4.0 amd64 [upgradable from: 2023.3.1]
	// libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 amd64 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]
	// libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 i386 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(string(msg), "\n")

	for _, line := range lines {
		// skip if line starts with "Listing..."
		if strings.HasPrefix(line, "Listing...") {
			continue
		}

		if len(line) > 0 {
			parts := strings.Fields(line)
			// log.Printf("apt: parts: %+v", parts)

			name := strings.Split(parts[0], "/")[0]
			category := strings.Split(parts[0], "/")[1]
			newVersion := parts[1]
			arch := parts[2]
			version := parts[5]
			version = strings.TrimSuffix(version, "]")

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        version,
				NewVersion:     newVersion,
				Category:       category,
				Arch:           arch,
				Status:         manager.PackageStatusUpgradable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// getPackageStatus takes a map of package names and manager.PackageInfo objects, and returns a list
// of manager.PackageInfo objects with their statuses updated using the output of `dpkg-query` command.
// It also adds any packages not found by dpkg-query to the list with their status set to unknown.
func getPackageStatus(packages map[string]manager.PackageInfo) ([]manager.PackageInfo, error) {
	var packageNames []string
	var packagesList []manager.PackageInfo

	if len(packages) == 0 {
		return packagesList, nil
	}

	for name := range packages {
		packageNames = append(packageNames, name)
	}

	args := []string{"-W", "--showformat", "${binary:Package} ${Status} ${Version}\n"}
	args = append(args, packageNames...)
	cmd := exec.Command("dpkg-query", args...)
	cmd.Env = ENV_NonInteractive

	// dpkg-query might exit with status 1, which is not an error when some packages are not found
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 1 && !strings.Contains(string(out), "no packages found matching") {
				return nil, fmt.Errorf("command failed with output: %s", string(out))
			}
		}
	}

	packagesList, err = ParseDpkgQueryOutput(out, packages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dpkg-query output: %+v", err)
	}

	// for all the packages that are not found, set their status to unknown, if any
	for _, pkg := range packages {
		fmt.Printf("apt: package not found by dpkg-query: %s", pkg.Name)
		pkg.Status = manager.PackageStatusUnknown
		packagesList = append(packagesList, pkg)
	}

	return packagesList, nil
}

// ParseDpkgQueryOutput parses the output of `dpkg-query` command and updates the status
// and version of the packages in the provided map of package names and manager.PackageInfo objects.
// It returns a list of manager.PackageInfo objects with their statuses and versions updated.
func ParseDpkgQueryOutput(output []byte, packages map[string]manager.PackageInfo) ([]manager.PackageInfo, error) {
	var packagesList []manager.PackageInfo

	// remove the last empty line
	output = bytes.TrimSuffix(output, []byte("\n"))
	lines := bytes.Split(output, []byte("\n"))

	for _, line := range lines {
		if len(line) > 0 {
			parts := bytes.Fields(line)

			if len(parts) < 2 {
				continue
			}

			name := string(parts[0])

			if strings.HasPrefix(name, "dpkg-query:") {
				name = string(parts[len(parts)-1])
			}

			if strings.Contains(name, ":") {
				name = strings.Split(name, ":")[0]
			}

			// if name is empty, it might not be what we want
			if name == "" {
				continue
			}

			version := string(parts[len(parts)-1])
			if !regexp.MustCompile(`^\d`).MatchString(version) {
				version = ""
			}

			pkg, ok := packages[name]

			if !ok {
				pkg = manager.PackageInfo{}
				packages[name] = pkg
			}

			delete(packages, name)

			switch {
			case bytes.HasPrefix(line, []byte("dpkg-query: ")):
				pkg.Status = manager.PackageStatusUnknown
				pkg.Version = ""
			case string(parts[len(parts)-2]) == "installed":
				pkg.Status = manager.PackageStatusInstalled
				if version != "" {
					pkg.Version = version
				}
			case string(parts[len(parts)-2]) == "config-files":
				pkg.Status = manager.PackageStatusConfigFiles
				if version != "" {
					pkg.Version = version
				}
			default:
				pkg.Status = manager.PackageStatusAvailable
				if version != "" {
					pkg.Version = version
				}
			}

			packagesList = append(packagesList, pkg)
		} else {
			log.Println("apt: line is empty")
		}
	}

	return packagesList, nil
}

// ParsePackageInfoOutput parses the output of `apt-cache show packageName` command
// and returns a manager.PackageInfo object containing package information such as name, version,
// architecture, and category. This function is useful for getting detailed package information.
func ParsePackageInfoOutput(msg string, opts *manager.Options) manager.PackageInfo {
	var pkg manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(string(msg), "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Package":
				pkg.Name = value
			case "Version":
				pkg.Version = value
			case "Architecture":
				pkg.Arch = value
			case "Section":
				pkg.Category = value
			}
		}
	}

	pkg.PackageManager = "apt"

	return pkg
}
