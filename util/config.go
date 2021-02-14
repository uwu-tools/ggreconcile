/*
Copyright 2021 The Kubernetes Authors.

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

package util

type Config struct {
	// the email id for the bot/service account
	BotID string `yaml:"bot-id"`

	// the gcloud secret containing a service account key to authenticate with
	SecretVersion string `yaml:"secret-version,omitempty"`

	// GroupsPath is the path to the directory with
	// groups.yaml files containing groups/members information.
	// It must be an absolute path. If not specified,
	// it defaults to the directory containing the config.yaml file.
	GroupsPath *string `yaml:"groups-path,omitempty"`

	// If false, don't make any mutating API calls
	ConfirmChanges bool
}
