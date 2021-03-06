/*
Copyright 2019 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

package pod

import (
	"fmt"
	"io"
	"os/exec"

	"k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo"
)

// RunInPod optionally tars up some files or directories, unpacks them in a container,
// and executes a shell command. Any error is treated as test failure.
func RunInPod(f *framework.Framework, rootdir string, items []string, command string, namespace, pod, container string) (string, string) {
	var input io.Reader
	var cmdPrefix string
	if len(items) > 0 {
		args := []string{"-cf", "-"}
		args = append(args, items...)
		tar := exec.Command("tar", args...)
		tar.Stderr = GinkgoWriter
		tar.Dir = rootdir
		pipe, err := tar.StdoutPipe()
		framework.ExpectNoError(err, "create pipe for tar")
		err = tar.Start()
		framework.ExpectNoError(err, "run tar")
		defer func() {
			err = tar.Wait()
			framework.ExpectNoError(err, "tar runtime error")
		}()

		input = pipe
		cmdPrefix = "tar -xf - &&"
	}

	options := framework.ExecOptions{
		Command: []string{
			"/bin/sh",
			"-c",
			cmdPrefix + command,
		},
		Namespace:     namespace,
		PodName:       pod,
		ContainerName: container,
		Stdin:         input,
		CaptureStdout: true,
		CaptureStderr: true,
	}
	stdout, stderr, err := f.ExecWithOptions(options)
	framework.ExpectNoError(err, "command failed in namespace %s, pod/container %s/%s:\nstderr:\n%s\nstdout:%s\n",
		namespace, pod, container, stderr, stdout)
	fmt.Fprintf(GinkgoWriter, "stderr:\n%s\nstdout:\n%s",
		stderr, stdout)

	return stdout, stderr
}
