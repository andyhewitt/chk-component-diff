/*
Copyright 2016 The Kubernetes Authors.
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

// Note: the example only works with the code within the same release/branch.
package main

import (
	chk_components "github.com/andyhewitt/chk-component-diff/pkg/chk-components-diff"
)

func main() {
	if chk_components.Labelarg != "" {
		chk_components.CompareLabels(chk_components.Labelarg, chk_components.Clustersarg)
	} else {
		chk_components.CompareComponents(chk_components.Resourcesarg, chk_components.Clustersarg)
	}
	// release_test.TestSum()
}
