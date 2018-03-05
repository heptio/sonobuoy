/*
Copyright 2018 Heptio Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"

	"github.com/heptio/sonobuoy/pkg/client"
	"github.com/heptio/sonobuoy/pkg/errlog"
)

var logConfig client.LogConfig
var logsKubecfg Kubeconfig

func init() {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Dumps the logs of the currently running sonobuoy containers for diagnostics",
		Run:   getLogs,
		Args:  cobra.ExactArgs(0),
	}

	logConfig.Follow = cmd.Flags().BoolP(
		"follow", "f", false,
		"Specify if the logs should be streamed.",
	)
	logConfig.Out = os.Stdout
	AddKubeconfigFlag(&logsKubecfg, cmd.Flags())
	AddNamespaceFlag(&logConfig.Namespace, cmd.Flags())
	RootCmd.AddCommand(cmd)
}

func getLogs(cmd *cobra.Command, args []string) {
	restConfig, err := logsKubecfg.Get()
	if err != nil {
		errlog.LogError(fmt.Errorf("failed to get rest config: %v", err))
		os.Exit(1)
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		errlog.LogError(fmt.Errorf("failed to get kubernetes client: %v", err))
		os.Exit(1)
	}

	errors := 0
	for err := range client.NewSonobuoyClient().StreamLogs(&logConfig, kubeClient) {
		errlog.LogError(err)
		errors++
	}

	if errors > 0 {
		os.Exit(1)
	}
}
