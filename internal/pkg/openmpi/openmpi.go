// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package openmpi

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/BTMichalowicz/go_exec/pkg/advexec"
	"github.com/BTMichalowicz/go_hpc_jobmgr/internal/pkg/network"
	"github.com/BTMichalowicz/go_hpc_jobmgr/pkg/sys"
	"github.com/BTMichalowicz/go_util/pkg/util"
)

const (
	// VersionTag is the tag used to refer to the MPI version in Open MPI template(s)
	VersionTag = "OMPIVERSION"

	// URLTag is the tag used to refer to the MPI URL in Open MPI template(s)
	URLTag = "OMPIURL"

	// TarballTag is the tag used to refer to the MPI tarball in Open MPI template(s)
	TarballTag = "OMPITARBALL"

	// ID is the internal ID for Open MPI
	ID = "openmpi"
)

// GetExtraMpirunArgs returns the set of arguments required for the mpirun command for the target platform
func GetExtraMpirunArgs(sys *sys.Config, netCfg *network.Config, extraArgs []string) []string {
	// By default we always prefer UCX rather than openib
	extraArgs = append(extraArgs, "--mca")
	extraArgs = append(extraArgs, "btl")
	extraArgs = append(extraArgs, "^openib")
	extraArgs = append(extraArgs, "--mca")
	extraArgs = append(extraArgs, "pml")
	extraArgs = append(extraArgs, "ucx")
	if netCfg != nil && netCfg.Device != "" {
		extraArgs = append(extraArgs, "-x UCX_NET_DEVICES="+netCfg.Device)
	}
	return extraArgs
}

func parseOmpiInfoOutputForVersion(output string) (string, error) {
	lines := strings.Split(output, "\n")
	if !strings.HasPrefix(lines[0], "Open MPI") {
		return "", fmt.Errorf("invalid output format")
	}
	version := strings.TrimPrefix(lines[0], "Open MPI v")
	version = strings.TrimRight(version, "\n")
	return version, nil
}

// DetectFromDir tries to figure out which version of OpenMPI is installed in a given directory
func DetectFromDir(dir string, env []string) (string, string, error) {
	targetBin := filepath.Join(dir, "bin", "ompi_info")
	if !util.FileExists(targetBin) {
		return "", "", fmt.Errorf("%s does not exist, not an OpenMPI implementation", targetBin)
	}

	var versionCmd advexec.Advcmd
	versionCmd.BinPath = targetBin
	versionCmd.CmdArgs = append(versionCmd.CmdArgs, "--version")
	versionCmd.ExecDir = filepath.Join(dir, "bin")
	versionCmd.Env = env
	if env == nil {
		newLDPath := filepath.Join(dir, "lib") + ":$LD_LIBRARY_PATH"
		newPath := filepath.Join(dir, "bin") + ":$PATH"
		versionCmd.Env = append(versionCmd.Env, "LD_LIBRARY_PATH="+newLDPath)
		versionCmd.Env = append(versionCmd.Env, "PATH="+newPath)
	}
	res := versionCmd.Run()
	if res.Err != nil {
		// If it fails we try with OPAL_PREFIX set. We create a new command to avoid "exec: already started" issues.
		var versionCmdWithOpalPrefix advexec.Advcmd
		versionCmdWithOpalPrefix.BinPath = versionCmd.BinPath
		versionCmdWithOpalPrefix.CmdArgs = versionCmd.CmdArgs
		versionCmdWithOpalPrefix.Env = versionCmd.Env
		versionCmdWithOpalPrefix.Env = append(versionCmd.Env, "OPAL_PREFIX="+dir)
		res = versionCmdWithOpalPrefix.Run()
		if res.Err != nil {
			log.Printf("unable to run ompi_info: %s; stdout: %s; stderr: %s", res.Err, res.Stdout, res.Stderr)
			return "", "", res.Err
		}
	}
	version, err := parseOmpiInfoOutputForVersion(res.Stdout)
	if err != nil {
		return "", "", fmt.Errorf("parseOmpiInfoOutputForVersion() failed - %w", err)
	}

	return ID, version, nil
}
