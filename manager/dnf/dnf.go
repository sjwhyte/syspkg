package dnf

import (
	"github.com/sjwhyte/syspkg/manager"
	"log"
	"os"
	"os/exec"
)

var pm string = "dnf"

// Constants used for dnf commands
const (
	ArgsAssumeYes      string = "-y"
	ArgsAssumeNo       string = "--assume-no"
	ArgsQuiet          string = "-q"
	ArgsPurge          string = "--purge"
	ArgsAutoRemove     string = "--autoremove"
	ArgsShowProgress   string = "--show-progress"
	ArgsShowDuplicates string = "--showduplicates"
)

// PackageManager implements the manager.PackageManager interface for the dnf package manager.
type PackageManager struct{}

// IsAvailable checks if the dnf package manager is available on the system.
func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

func (a *PackageManager) Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"search"}, ArgsShowDuplicates)
	args = append(args, keywords...)

	cmd := exec.Command("dnf", args...)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return ParseFindOutput(string(out), true, opts), nil
}

func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command("dnf", "list", "installed", "${binary:Package} ${Version}\n")
	// NOTE: can also use `apt list --installed`, but it's slower
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseInstallOutput(string(out), opts), nil
}

func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	//TODO implement me
	panic("implement me")
}

// Upgrade upgrades the provided packages using the apt package manager.
func (a *PackageManager) Upgrade(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"upgrade"}
	if len(pkgs) > 0 {
		args = append(args, pkgs...)
	}

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	cmd := exec.Command(pm, args...)

	log.Printf("Running command: %s %s", pm, args)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseInstallOutput(string(out), opts), nil
}

func (a *PackageManager) UpgradeAll(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	return a.Upgrade(pkgs, opts)
}

func (a *PackageManager) GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error) {
	err := a.Refresh(nil)
	if err != nil {
		return manager.PackageInfo{}, err
	}
	cmd := exec.Command("info", pkg)

	out, err := cmd.Output()
	if err != nil {
		return manager.PackageInfo{}, err
	}
	return ParsePackageInfoOutput(string(out), opts), nil
}

// GetPackageManager returns the name of the dnf package manager.
func (a *PackageManager) GetPackageManager() string {
	return pm
}

// Install installs the provided packages using the apt package manager.
func (a *PackageManager) Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"install"}, pkgs...)

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	// assume yes if not interactive, to avoid hanging
	if !opts.Interactive {
		args = append(args, ArgsAssumeYes)
	}

	cmd := exec.Command(pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	} else {
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return ParseInstallOutput(string(out), opts), nil
	}
}

// Delete removes the provided packages using the apt package manager.
func (a *PackageManager) Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// args := append([]string{"remove", ArgsFixBroken, ArgsPurge, ArgsAutoRemove}, pkgs...)
	args := append([]string{"remove", ArgsAutoRemove}, pkgs...)
	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	if !opts.Interactive {
		args = append(args, ArgsAssumeYes)
	}

	cmd := exec.Command(pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	} else {
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return ParseDeletedOutput(string(out), opts), nil
	}
}

// Refresh updates the package list using the apt package manager.
func (a *PackageManager) Refresh(opts *manager.Options) error {
	cmd := exec.Command(pm, "update")

	if opts == nil {
		opts = &manager.Options{
			Verbose:   false,
			AssumeYes: true,
		}
	}
	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return err
	} else {
		out, err := cmd.Output()
		if err != nil {
			return err
		}
		if opts.Verbose {
			log.Println(string(out))
		}
		return nil
	}
}
