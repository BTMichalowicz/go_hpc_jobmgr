// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package jm

import (
	"testing"

	"github.com/BTMichalowicz/go_hpc_jobmgr/pkg/job"
	"github.com/BTMichalowicz/go_hpc_jobmgr/pkg/sys"
	"github.com/BTMichalowicz/go_util/pkg/util"
)

func TestDetect(t *testing.T) {
	jm := Detect()
	t.Logf("Selected job manager: %s\n", jm.ID)
}

func TestTempFile(t *testing.T) {
	var j job.Job
	var sysCfg sys.Config

	err := TempFile(&j, &sysCfg)
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	if j.BatchScript == "" {
		t.Fatalf("temporary file path is undefined")
	}

	t.Logf("Temporary file is: %s\n", j.BatchScript)
	err = j.CleanUp()
	if err != nil {
		t.Fatalf("failed to clean up: %s", err)
	}

	if util.PathExists(j.BatchScript) {
		t.Fatalf("temporary file %s still exists even after cleanup", j.BatchScript)
	}
}
