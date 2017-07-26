/*
Copyright 2017 Heptio Inc.

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

package loader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/plugin"
	"github.com/heptio/sonobuoy/pkg/plugin/driver/daemonset"
	"github.com/heptio/sonobuoy/pkg/plugin/driver/job"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
)

// LoadAllPlugins loads all plugins by finding plugin definitions in the given
// directory, taking a user's plugin selections, and a sonobuoy phone home
// address (host:port) and returning all of the active, configured plugins for
// this sonobuoy run.
func LoadAllPlugins(namespace string, searchPath []string, selections []plugin.Selection, masterAddress string) (ret []plugin.Interface, err error) {
	var defns []plugin.Definition

	for _, dir := range searchPath {
		wd, _ := os.Getwd()
		glog.Infof("Scanning plugins in %v (pwd: %v)", dir, wd)

		// We only care about configured plugin directories that exist,
		// since we may have a broad search path.
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		p, err := scanPlugins(dir)
		if err != nil {
			return ret, err
		}

		defns = append(defns, p...)
	}

	for _, selection := range selections {
		for _, pluginDef := range defns {
			if selection.Name == pluginDef.Name {
				p, err := loadPlugin(namespace, pluginDef, masterAddress)
				if err != nil {
					return ret, err
				}
				ret = append(ret, p)
			}
		}
	}
	return ret, nil
}

// loadPlugin loads an individual plugin by instantiating a plugin driver with
// the settings from the given plugin definition and selection
func loadPlugin(namespace string, dfn plugin.Definition, masterAddress string) (plugin.Interface, error) {
	cfg := &plugin.WorkerConfig{
		ResultType: dfn.ResultType,
	}

	switch dfn.Driver {
	case "DaemonSet":
		cfg.MasterURL = "http://" + masterAddress + "/api/v1/results/by-node"
		return daemonset.NewPlugin(namespace, dfn, cfg), nil
	case "Job":
		cfg.MasterURL = "http://" + masterAddress + "/api/v1/results/global"
		return job.NewPlugin(namespace, dfn, cfg), nil
	default:
		return nil, fmt.Errorf("Unknown driver %v", dfn.Driver)
	}
}

// scanPlugins looks for Plugin Definition YAML files in the given directory,
// and returns an array of PluginDefinition structs.
func scanPlugins(dir string) ([]plugin.Definition, error) {
	var plugins []plugin.Definition

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return plugins, err
	}

	for _, file := range files {
		// We only look at .yaml files in this directory
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		// Read the file into memory
		fullPath := path.Join(dir, file.Name())
		y, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return plugins, err
		}

		// Load it into a proper PluginDefinition.  If we can't, just
		// warn.  If they've selected this plugin in their config,
		// they'll get an error then.
		pluginDef, err := loadPluginDefinition(y)
		if err != nil {
			glog.Warningf("Error loading plugin at %v: %v", fullPath, err)
			continue
		}

		plugins = append(plugins, *pluginDef)
	}

	return plugins, err
}

// loadPluginDefinition takes a YAML string of bytes and loads a
// plugin.Definition.
func loadPluginDefinition(pluginYaml []byte) (*plugin.Definition, error) {
	var ret plugin.Definition

	err := yaml.Unmarshal(pluginYaml, &ret)
	if err != nil {
		return nil, err
	}

	// Validate it
	if ret.Driver == "" {
		return nil, fmt.Errorf("No driver specified in plugin YAML")
	}
	if ret.ResultType == "" {
		return nil, fmt.Errorf("No resultType specified in plugin YAML")
	}
	if ret.Name == "" {
		return nil, fmt.Errorf("No name specified in plugin YAML")
	}
	if ret.RawPodSpec == nil {
		return nil, fmt.Errorf("No pod spec specified in plugin YAML")
	}

	// Construct a pod spec from the YAML. We can't decode it
	// directly since a PodSpec is not a runtime.Object (it doesn't
	// have ObjectMeta attributes like Kind and Metadata), so we:

	// make a fake pod as a map[string]interface{}, and load the
	// plugin config yaml into its spec
	placeholderPodMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"spec":       ret.RawPodSpec,
	}

	// serialize the result into YAML
	placeholderPodYaml, err := yaml.Marshal(placeholderPodMap)
	if err != nil {
		return nil, err
	}

	// Decode *that* yaml into a Pod
	var placeholderPod v1.Pod
	if err := kuberuntime.DecodeInto(scheme.Codecs.UniversalDecoder(), placeholderPodYaml, &placeholderPod); err != nil {
		glog.Fatalf("Could not decode pod spec: %v", err)
	}
	ret.PodSpec = placeholderPod.Spec

	return &ret, nil
}
