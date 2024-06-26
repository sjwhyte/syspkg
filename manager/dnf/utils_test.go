package dnf_test

import (
	"github.com/bluet/syspkg/manager/dnf"
	"testing"
)

var findOutput = `============================================================================================================================================================================= Name Exactly Matched: corelight-sensor ==============================================================================================================================================================================
corelight-sensor-27.10.0-1.x86_64 : Corelight Next-Gen Sensor
corelight-sensor-27.11.0-1.x86_64 : Corelight Next-Gen Sensor
corelight-sensor-27.11.1-1.x86_64 : Corelight Next-Gen Sensor
corelight-sensor-27.11.2-1.x86_64 : Corelight Next-Gen Sensor
corelight-sensor-27.9.0-1.x86_64 : Corelight Next-Gen Sensor
corelight-sensor-27.9.1-1.x86_64 : Corelight Next-Gen Sensor
corelight-sensor-27.9.2-1.x86_64 : Corelight Next-Gen Sensor
================================================================================================================================================================================= Name Matched: corelight-sensor ==================================================================================================================================================================================
corelight-sensor-images-base-27.10.0-1.x86_64 : Corelight Next-Gen Sensor Images
corelight-sensor-images-base-27.11.0-1.x86_64 : Corelight Next-Gen Sensor Images
corelight-sensor-images-base-27.11.1-1.x86_64 : Corelight Next-Gen Sensor Images
corelight-sensor-images-base-27.11.2-1.x86_64 : Corelight Next-Gen Sensor Images
corelight-sensor-images-base-27.9.0-1.x86_64 : Corelight Next-Gen Sensor Images
corelight-sensor-images-base-27.9.1-1.x86_64 : Corelight Next-Gen Sensor Images
corelight-sensor-images-base-27.9.2-1.x86_64 : Corelight Next-Gen Sensor Images
corelight-sensor-images-ml-27.10.0-1.x86_64 : Corelight Sensor ML Images
corelight-sensor-images-ml-27.11.0-1.x86_64 : Corelight Sensor ML Images
corelight-sensor-images-ml-27.11.1-1.x86_64 : Corelight Sensor ML Images
corelight-sensor-images-ml-27.11.2-1.x86_64 : Corelight Sensor ML Images
corelight-sensor-images-ml-27.9.0-1.x86_64 : Corelight Sensor ML Images
corelight-sensor-images-ml-27.9.1-1.x86_64 : Corelight Sensor ML Images
corelight-sensor-images-ml-27.9.2-1.x86_64 : Corelight Sensor ML Images
corelight-sensor-images-rke2-27.10.0-1.x86_64 : Corelight Sensor RKE2 Images
corelight-sensor-images-rke2-27.11.0-1.x86_64 : Corelight Sensor RKE2 Images
corelight-sensor-images-rke2-27.11.1-1.x86_64 : Corelight Sensor RKE2 Images
corelight-sensor-images-rke2-27.11.2-1.x86_64 : Corelight Sensor RKE2 Images
corelight-sensor-images-rke2-27.9.0-1.x86_64 : Corelight Sensor RKE2 Images
corelight-sensor-images-rke2-27.9.1-1.x86_64 : Corelight Sensor RKE2 Images
corelight-sensor-images-rke2-27.9.2-1.x86_64 : Corelight Sensor RKE2 Images
`

var installedOutput = `Installed Packages
NetworkManager.x86_64                                                                                                                                                                1:1.40.16-15.el8_9                                                                                                                                                                           @baseos
NetworkManager-cloud-setup.x86_64                                                                                                                                                    1:1.40.16-15.el8_9                                                                                                                                                                           @appstream
NetworkManager-libnm.x86_64                                                                                                                                                          1:1.40.16-15.el8_9                                                                                                                                                                           @baseos
NetworkManager-team.x86_64                                                                                                                                                           1:1.40.16-15.el8_9                                                                                                                                                                           @baseos
NetworkManager-tui.x86_64                                                                                                                                                            1:1.40.16-15.el8_9                                                                                                                                                                           @baseos
acl.x86_64                                                                                                                                                                           2.2.53-3.el8                                                                                                                                                                                 @baseos
almalinux-release.x86_64                                                                                                                                                             8.10-1.el8                                                                                                                                                                                   @baseos
amazon-ssm-agent.x86_64                                                                                                                                                              3.2.1798.0-1                                                                                                                                                                                 @System
audit.x86_64                                                                                                                                                                         3.1.2-1.el8                                                                                                                                                                                  @baseos
audit-libs.x86_64                                                                                                                                                                    3.1.2-1.el8                                                                                                                                                                                  @baseos
authselect.x86_64                                                                                                                                                                    1.2.6-2.el8                                                                                                                                                                                  @System
authselect-libs.x86_64                                                                                                                                                               1.2.6-2.el8                                                                                                                                                                                  @System
basesystem.noarch                                                                                                                                                                    11-5.el8                                                                                                                                                                                     @System
bash.x86_64                                                                                                                                                                          4.4.20-5.el8                                                                                                                                                                                 @baseos
bind-export-libs.x86_64                                                                                                                                                              32:9.11.36-14.el8_10                                                                                                                                                                         @baseos
binutils.x86_64                                                                                                                                                                      2.30-123.el8                                                                                                                                                                                 @baseos
`

var packageInfo = `Installed Packages
Name         : gzip
Version      : 1.9
Release      : 13.el8_5
Architecture : x86_64
Size         : 345 k
Source       : gzip-1.9-13.el8_5.src.rpm
Repository   : @System
Summary      : The GNU data compression program
URL          : http://www.gzip.org/
License      : GPLv3+ and GFDL
Description  : The gzip package contains the popular GNU gzip data compression
             : program. Gzipped files have a .gz extension.
             : 
             : Gzip should be installed on your system, because it is a
             : very commonly used data compression program.
`

func TestParseFindOutputExact(t *testing.T) {
	packageInfos := dnf.ParseFindOutput(findOutput, true, nil)

	if len(packageInfos) != 7 {
		t.Errorf("should have returned 7 lines, but got %v", len(packageInfos))
	}
}

func TestParseFindOutput(t *testing.T) {
	packageInfos := dnf.ParseFindOutput(findOutput, false, nil)

	if len(packageInfos) != 28 {
		t.Errorf("should have returned 28 lines, but got %v", len(packageInfos))
	}
}

func TestParseInstalledOuput(t *testing.T) {
	packageInfos := dnf.ParseInstallOutput(installedOutput, nil)
	if len(packageInfos) != 16 {
		t.Errorf("should have returned 16 lines, but got %v", len(packageInfos))
	}
}

func TestParsePackageInfoOutput(t *testing.T) {
	packageInfo := dnf.ParsePackageInfoOutput(packageInfo, nil)

	if packageInfo.Name != "gzip" {
		t.Errorf("should have returned gzip name, but got %v", packageInfo.Name)
	}
}
