// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package jm

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/BTMichalowicz/go_exec/pkg/advexec"
	"github.com/BTMichalowicz/go_hpc_jobmgr/internal/pkg/network"
	"github.com/BTMichalowicz/go_hpc_jobmgr/pkg/job"
	"github.com/BTMichalowicz/go_hpc_jobmgr/pkg/mpi"
	"github.com/BTMichalowicz/go_hpc_jobmgr/pkg/sys"
)

// Native is the structure representing the native job manager (i.e., directly use mpirun)
type Native struct {
}

// nativeGetOutput retrieves the application's output after the completion of a job
func nativeGetOutput(j *job.Job, sysCfg *sys.Config) string {
	return j.OutBuffer.String()
}

// nativeGetError retrieves the error messages from an application after the completion of a job
func nativeGetError(j *job.Job, sysCfg *sys.Config) string {
	return j.ErrBuffer.String()
}

func prepareMPISubmit(cmd *advexec.Advcmd, j *job.Job, sysCfg *sys.Config, netCfg *network.Config) error {
	var err error
	cmd.BinPath = filepath.Join(j.MPICfg.Implem.InstallDir, "bin", "mpirun")
	if j.NP > 0 {
		cmd.CmdArgs = append(cmd.CmdArgs, "-np")
		cmd.CmdArgs = append(cmd.CmdArgs, strconv.Itoa(j.NP))
	}

	mpirunArgs, err := mpi.GetMpirunArgs(&j.MPICfg.Implem, &j.App, sysCfg, netCfg, j.MPICfg.UserMpirunArgs)
	if err != nil {
		return fmt.Errorf("unable to get mpirun arguments: %s", err)
	}
	if len(mpirunArgs) > 0 {
		cmd.CmdArgs = append(cmd.CmdArgs, mpirunArgs...)
	}

	//newPath := getEnvPath(j.HostCfg, env)
	//newLDPath := getEnvLDPath(j.HostCfg, env)
	//log.Printf("-> PATH=%s", newPath)
	//log.Printf("-> LD_LIBRARY_PATH=%s\n", newLDPath)
	//log.Printf("Using %s as PATH\n", newPath)
	//log.Printf("Using %s as LD_LIBRARY_PATH\n", newLDPath)
	//cmd.Env = append([]string{"LD_LIBRARY_PATH=" + newLDPath}, os.Environ()...)
	//cmd.Env = append([]string{"PATH=" + newPath}, os.Environ()...)

	return nil
}

/*
func prepareStdSubmit(cmd *advexec.Advcmd, j *job.Job, env *Environment, sysCfg *sys.Config) error {
	cmd.BinPath = j.App.BinPath
	cmd.CmdArgs = append(cmd.CmdArgs, j.App.BinArgs...)

	return nil
}
*/

// nativeSubmit is the function to call to submit a job through the native job manager
func nativeSubmit(j *job.Job, jobmgr *JM, sysCfg *sys.Config) advexec.Result {
	var cmd advexec.Advcmd
	var res advexec.Result

	if j.App.BinPath == "" {
		res.Err = fmt.Errorf("application binary is undefined")
		return res
	}

	netCfg := new(network.Config)
	netCfg.Device = j.Device

	err := prepareMPISubmit(&cmd, j, sysCfg, netCfg)
	if err != nil {
		res.Err = fmt.Errorf("unable to prepare MPI job: %s", err)
		return res
	}

	j.SetOutputFn(nativeGetOutput)
	j.SetErrorFn(nativeGetError)

	if j.RunDir != "" {
		cmd.ExecDir = j.RunDir
	}
	return cmd.Run()
}

func nativeLoad(jobmgr *JM, sysCfg *sys.Config) error {
	return nil
}

// NativeDetect is the function used by our job management framework to figure out if mpirun should be used directly.
// The native component is the default job manager. If application, the function returns a structure with all the
// "function pointers" to correctly use the native job manager.
func NativeDetect() (bool, JM) {
	var jm JM
	jm.ID = NativeID
	jm.submitJM = nativeSubmit
	jm.loadJM = nativeLoad
	jm.jobStatusJM = nil // Not implemented yet
	jm.postRunJM = nil   // Not implemented yet

	// This is the default job manager, i.e., mpirun so we do not check anything, just return this component.
	// If the component is selected and mpirun not correctly installed, the framework will pick it up later.
	return true, jm
}
