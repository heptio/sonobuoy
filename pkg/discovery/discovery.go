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

package discovery

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/config"
	pluginaggregation "github.com/heptio/sonobuoy/pkg/plugin/aggregation"
	"github.com/viniciuschiele/tarx"
	"k8s.io/client-go/kubernetes"
)

// Run is the main entrypoint for discovery
func Run(kubeClient kubernetes.Interface, cfg *config.Config) []error {
	var errlst []error

	t := time.Now()
	// 1. Get the list of namespaces and apply the regex filter on the namespace
	nslist := FilterNamespaces(kubeClient, cfg.Filters.Namespaces)

	// 2. Create the directory which will store the results
	outpath := cfg.ResultsDir + "/" + cfg.UUID
	err := os.MkdirAll(outpath, 0755)
	if err != nil {
		panic(err.Error())
	}

	// 3. Dump the config.json we used to run our test
	if blob, err := json.Marshal(cfg); err == nil {
		if err = ioutil.WriteFile(outpath+"/config.json", blob, 0644); err != nil {
			panic(err.Error())
		}
	}

	// closure used to collect and report errors.
	rollup := func(err []error) {
		if err != nil {
			errlst = append(errlst, err...)
		}
	}

	// 4. Run the plugin aggregator
	errlst = append(errlst, pluginaggregation.Run(kubeClient, cfg.LoadedPlugins, cfg.Aggregation, outpath)...)

	// 5. Run the queries
	rollup(QueryClusterResources(kubeClient, cfg))
	for _, ns := range nslist {
		rollup(QueryNSResources(kubeClient, ns, cfg))
	}

	// 6. Clean up after the plugins
	errlst = append(errlst, pluginaggregation.Cleanup(kubeClient, cfg.LoadedPlugins)...)

	// 7. tarball up results YYYYMMDDHHMM_sonobuoy_UID.tar.gz
	tb := cfg.ResultsDir + "/" + t.Format("200601021504") + "_sonobuoy_" + cfg.UUID + ".tar.gz"
	err = tarx.Compress(tb, outpath, &tarx.CompressOptions{Compression: tarx.Gzip})
	if err == nil {
		err = os.RemoveAll(outpath)
	}
	if err != nil {
		errlst = append(errlst, err)
	}

	glog.Infof("Results available at %v", tb)
	return errlst
}
