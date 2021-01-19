// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package jm

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
)

// Prun is the structure representing the native job manager (i.e., directly use mpirun)
type Prun struct {
}

// PrunGetOutput retrieves the application's output after the completion of a job
func PrunGetOutput(j *job.Job, sysCfg *sys.Config) string {
	return j.OutBuffer.String()
}

// PrunGetError retrieves the error messages from an application after the completion of a job
func PrunGetError(j *job.Job, sysCfg *sys.Config) string {
	return j.ErrBuffer.String()
}

// PrunSubmit is the function to call to submit a job through the native job manager
func PrunSubmit(j *job.Job, sysCfg *sys.Config) (advexec.Advcmd, error) {
	var cmd advexec.Advcmd
	var err error

	if j.App.BinPath == "" {
		return cmd, fmt.Errorf("application binary is undefined")
	}

	cmd.BinPath, err = exec.LookPath("prun")
	if err != nil {
		return cmd, fmt.Errorf("prun not found")
	}

	for _, a := range j.Args {
		cmd.CmdArgs = append(cmd.CmdArgs, a)
	}
	cmd.CmdArgs = append(cmd.CmdArgs, "-x")
	cmd.CmdArgs = append(cmd.CmdArgs, "PATH")
	cmd.CmdArgs = append(cmd.CmdArgs, j.App.BinPath)
	cmd.CmdArgs = append(cmd.CmdArgs, j.App.BinArgs...)

	//newPath := getEnvPath(j.HostCfg, env)
	//newLDPath := getEnvLDPath(j.HostCfg, env)
	//log.Printf("-> PATH=%s", newPath)
	//log.Printf("-> LD_LIBRARY_PATH=%s\n", newLDPath)
	//log.Printf("Using %s as PATH\n", newPath)
	//log.Printf("Using %s as LD_LIBRARY_PATH\n", newLDPath)
	//cmd.Env = append([]string{"LD_LIBRARY_PATH=" + newLDPath}, os.Environ()...)
	//cmd.Env = append([]string{"PATH=" + newPath}, sycmd.Env...)

	j.GetOutput = PrunGetOutput
	j.GetError = PrunGetError

	return cmd, nil
}

// PrunDetect is the function used by our job management framework to figure out if mpirun should be used directly.
// The native component is the default job manager. If application, the function returns a structure with all the
// "function pointers" to correctly use the native job manager.
func PrunDetect() (bool, JM) {
	var jm JM

	_, err := exec.LookPath("prun")
	if err != nil {
		log.Println("* prun not detected")
		return false, jm
	}

	jm.ID = PrunID
	jm.Submit = PrunSubmit

	// This is the default job manager, i.e., mpirun so we do not check anything, just return this component.
	// If the component is selected and mpirun not correctly installed, the framework will pick it up later.
	return true, jm
}
