// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2020-2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package jm

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/BTMichalowicz/go_exec/pkg/advexec"
	"github.com/BTMichalowicz/go_hpc_jobmgr/pkg/job"
	"github.com/BTMichalowicz/go_hpc_jobmgr/pkg/sys"
	"github.com/BTMichalowicz/go_util/pkg/util"
)

// IntelSlurmDetect is the function used by our job management framework to figure out if Intel-Slurm can be used and
// if so return a JM structure with all the "function pointers" to interact with Slurm through our generic
// API.
func IntelSlurmDetect() (bool, JM) {
	var jm JM
	var err error

	jm.BinPath, err = exec.LookPath("bsub")
	if err != nil {
		log.Println("* Intel-Slurm not detected")
		return false, jm
	}

	_, err = exec.LookPath("squeue")
	if err != nil {
		log.Println("* Intel-Slurm not detected (no squeue command available)")
		return false, jm
	}

	jm.ID = IntelSlurmID
	jm.submitJM = intelSlurmSubmit
	jm.loadJM = intelSlurmLoad
	jm.jobStatusJM = slurmJobStatus
	jm.numJobsJM = slurmGetNumJobs
	jm.postRunJM = slurmPostJob

	return true, jm
}

// intelSlurmLoad is the function called when trying to load a JM module
func intelSlurmLoad(jobmgr *JM, sysCfg *sys.Config) error {
	// jobmgr.BinPath has been set during Detect()
	return nil
}

// intelSlurmSubmit prepares the batch script necessary to start a given job.
//
// Note that a script does not need any specific environment to be submitted
func intelSlurmSubmit(j *job.Job, jobmgr *JM, sysCfg *sys.Config) advexec.Result {
	var cmd advexec.Advcmd
	var resExec advexec.Result

	// Sanity checks
	if j == nil || !util.FileExists(jobmgr.BinPath) {
		resExec.Err = fmt.Errorf("job is undefined")
		return resExec
	}

	err := generateJobScript(j, sysCfg)
	if err != nil {
		resExec.Err = fmt.Errorf("unable to generate Slurm script: %s", err)
		return resExec
	}
	if j.BatchScript == "" {
		resExec.Err = fmt.Errorf("undefined batch script path")
		return resExec
	}

	cmd.BinPath = jobmgr.BinPath
	cmd.ExecDir = j.RunDir
	// We want the default to be blocking sbatch but users can request non-blocking
	if !j.NonBlocking {
		jobmgr.CmdArgs = append(jobmgr.CmdArgs, "-W")
	}

	if len(jobmgr.CmdArgs) > 0 {
		cmd.CmdArgs = append(cmd.CmdArgs, jobmgr.CmdArgs...)
	}
	//cmd.CmdArgs = append(cmd.CmdArgs, j.BatchScript)
	cmd.CmdArgs = []string{j.BatchScript}

	j.SetOutputFn(slurmGetOutput)
	j.SetErrorFn(slurmGetError)

	if !util.PathExists(sysCfg.ScratchDir) {
		resExec.Err = fmt.Errorf("scratch directory does not exist")
		return resExec
	}

	cmdRes := cmd.Run()
	if strings.HasPrefix(cmdRes.Stdout, slurmJobIDPrefix) {
		jobIDStr := strings.TrimPrefix(cmdRes.Stdout, slurmJobIDPrefix)
		jobIDStr = strings.TrimRight(jobIDStr, "\n")
		j.ID, err = strconv.Atoi(jobIDStr)
		if err != nil {
			resExec.Err = fmt.Errorf("unable to get job ID: %s", err)
			return resExec
		}
	}

	if !j.NonBlocking {
		return slurmPostJob(&cmdRes, j, sysCfg)
	}

	return cmdRes
}
