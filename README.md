# Sonobuoy

**Maintainers:** [Heptio][0]

[![Build Status][1]][2]


## Overview

Sonobuoy addresses the gap in tools for [Kubernetes][3] cluster introspection. It is a customizable, extendable, and cluster-agnostic way to record snapshots of your cluster's essential characteristics.

Sonobuoy generates clear, informative reports about your cluster, regardless of your deployment details. Its selective data dumps of Kubernetes resource objects and cluster nodes allow for the following use cases:

* Integrated end-to-end (e2e) [conformance-testing 1.7+][13]
* Workload debugging
* Workload recovery
* Custom data collection via extensible plugins

## Install and Configure

Heptio provides prebuilt Sonobuoy container images in its Google Container Registry (*gcr.io/heptio-images*). For the sake of faster setup on your cluster, **this quickstart pulls from this registry to skip the container build process**. You can use this same process to deploy Sonobuoy to production.

See [Build From Scratch][4] for instructions on building Sonobuoy yourself.


### 0. Prerequisites

* *You should have access to an up-and-running Kubernetes cluster.* If you do not have a cluster, follow the [AWS Quickstart Kubernetes Tutorial][5] to set one up with a single command.

* *You should have `kubectl` installed.* If not, follow the instructions for [installing via Homebrew (MacOS)][6] or [building the binary (Linux)][7].


### 1. Download
Clone or fork the Sonobuoy repo:
```
git clone git@github.com:heptio/sonobuoy.git
```


### 2. Run

Run the following command in the Sonobuoy root directory to deploy a Sonobuoy pod to your cluster:
```
kubectl apply -f examples/quickstart/
```

You can view actively running pods with the following command:
```
kubectl get pods -l component=sonobuoy --namespace=heptio-sonobuoy
```

To verify that Sonobuoy has completed successfuly, check the logs:
```
kubectl logs -f sonobuoy --namespace=heptio-sonobuoy
```
If you see the log line `no-exit was specified, sonobuoy is now blocking`, the Sonobuoy pod is done running.

To view the output, copy the output directory from the main Sonobuoy pod to somewhere local:
```
kubectl cp heptio-sonobuoy/sonobuoy:/tmp/sonobuoy ./results --namespace=heptio-sonobuoy
```

There should a collection of tarballs inside of `./results` , where each tarball corresponds to a single Sonobuoy run. If you unzip one of these data dumps, you should see sub-directories containing info about `hosts`, `plugins`, `resources`, and `serverversion`. If you have time, look through these directories to get a sense for Sonobuoy's capabilities.

*NOTE: At this time, the layout of the contents of the tarball is subject to change.*

### 3. Tear down

Sonobuoy is not a persistent, background process---each time you want a new data report you will need to re-run it.

To clean up Kubernetes objects created by Sonobuoy, run the following commands:
```
kubectl delete -f examples/quickstart/
```
You may also want to clear the contents of the results directory that the Sonobuoy pod wrote to.


## Further documentation

 To learn how to configure your Sonobuoy runs and integrate plugins, see the [`/docs` directory][9].

## Troubleshooting

If you encounter any problems that the documentation does not address, [file an issue][10].  

## Contributing

Thanks for taking the time to join our community and start contributing!

#### Before you start

* Please familiarize yourself with the [Code of
Conduct][12] before contributing.
* See [CONTRIBUTING.md][11] for instructions on the
developer certificate of origin that we require.

#### Pull requests

* We welcome pull requests. Feel free to dig through the [issues][10] and jump in.


## Changelog

See [the list of releases](/CHANGELOG.md) to find out about feature changes.

[0]: https://github.com/heptio
[1]: https://jenkins.i.heptio.com/buildStatus/icon?job=sonobuoy-master-deployer
[2]: https://jenkins.i.heptio.com/job/sonobuoy-master-deployer/
[3]: https://github.com/kubernetes/kubernetes
[4]: /docs/build-from-scratch.md
[5]: http://docs.heptio.com/content/tutorials/aws-cloudformation-k8s.html
[6]: https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-with-homebrew-on-macos
[7]: https://kubernetes.io/docs/tasks/tools/install-kubectl/#tabset-1
[8]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/
[9]: /docs
[10]: https://github.com/heptio/sonobuoy/issues
[11]: /CONTRIBUTING.md
[12]: /CODE_OF_CONDUCT.md
[13]: /docs/conformance-testing.md
