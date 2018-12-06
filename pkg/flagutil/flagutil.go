/*
Copyright 2018 The Kubernetes Authors.

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

// Package flagutil contains utilities and interfaces shared between
// several test-infra commands.
package flagutil

import (
	"flag"
)

// OptionGroup provides an interface which can be implemented by an
// option handler (e.g. for GitHub or Kubernetes) to support generic
// option-group handling.
type OptionGroup interface {
	// AddFlags injects options into the given FlagSet.
	AddFlags(fs *flag.FlagSet)

	// Validate validates options.
	Validate(dryRun bool) error
}
