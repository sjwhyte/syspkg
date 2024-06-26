package dnf

import (
	"github.com/bluet/syspkg/manager"
	"log"
	"regexp"
	"strings"
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

	re := regexp.MustCompile(`^(\S+)\.(\S+)\s+(\S+)\s+@(\S+)`)

	lines := strings.Split(msg, "\n")

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			pkgInfo := manager.PackageInfo{
				Name:           matches[1],
				Version:        matches[3],
				Arch:           matches[2],
				Category:       matches[4],
				PackageManager: pm,
			}
			packages = append(packages, pkgInfo)
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
func ParseFindOutput(msg string, exactMatch bool, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	var lines []string
	sections := strings.Split(msg, "=============================================================================================================================================================================")
	exactMatches := strings.Split(sections[2], "\n")
	nameMatches := strings.Split(sections[4], "\n")

	packageMatches := make([]string, 0)

	for _, line := range exactMatches {
		line = strings.TrimSpace(line)
		if line != "" && (!strings.HasPrefix(line, "Name Exactly Matched:") && !strings.HasPrefix(line, "=")) {
			packageMatches = append(packageMatches, line)
		}
	}

	if !exactMatch {
		for _, line := range nameMatches {
			line = strings.TrimSpace(line)
			if line != "" && (!strings.HasPrefix(line, "Name Matched:") && !strings.HasPrefix(line, "=")) {
				packageMatches = append(packageMatches, line)
			}
		}
	}

	for _, line := range packageMatches {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "=") {
			lines = append(lines, line)
		}
	}

	for _, line := range lines {
		pi := parsePackageInfo(line)

		packages = append(packages, pi)
	}

	return packages
}

func parsePackageInfo(input string) (packages manager.PackageInfo) {
	// Define the regex pattern
	pattern := `^(?P<packageName>.+)-(?P<version>\d+\.\d+\.\d+-\d+)\.(?P<architecture>.+)$`
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(strings.TrimRight(strings.SplitN(input, ":", 2)[0], " "))
	if match == nil {
		return manager.PackageInfo{}
	}

	// Extract the named capture groups
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	return manager.PackageInfo{
		Name:           result["packageName"],
		Version:        result["version"],
		NewVersion:     result["version"],
		Arch:           result["architecture"],
		PackageManager: pm,
	}
}

func ParsePackageInfoOutput(msg string, opts *manager.Options) manager.PackageInfo {

	patterns := map[string]*regexp.Regexp{
		"Name":         regexp.MustCompile(`Name\s+:\s+(.+)`),
		"Version":      regexp.MustCompile(`Version\s+:\s+(.+)`),
		"Release":      regexp.MustCompile(`Release\s+:\s+(.+)`),
		"Architecture": regexp.MustCompile(`Architecture\s+:\s+(.+)`),
	}

	lines := strings.Split(msg, "\n")
	pi := manager.PackageInfo{
		PackageManager: pm,
	}
	for _, line := range lines {
		for field, pattern := range patterns {
			if match := pattern.FindStringSubmatch(line); len(match) > 1 {
				switch field {
				case "Name":
					pi.Name = match[1]
				case "Version":
					pi.Version = match[1]
				case "Release":
					pi.Version = pi.Version + "-" + match[1]
				case "Architecture":
					pi.Arch = match[1]
				}
			}
		}
	}

	return pi
}
