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

// Package buildinfo holds build-time information like the sonobuoy version.
// This is a separate package so that other packages can import it without
// worrying about introducing circular dependencies.
package buildinfo

// Version is the current version of Sonobuoy, set by the go linker's -X flag at build time
var Version string

// DockerImage is the full path to the docker image for this build, example gcr.io/heptio-images/sonobuoy
var DockerImage string
