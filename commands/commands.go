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

package commands

import (
	"fmt"

	base "github.com/n3wscott/cli-base/pkg/commands/options"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/log"
)

var opts = &Options{}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ggreconcile",
		Short:             base.Wrap80("Google Groups reconciler."),
		PersistentPreRunE: initLogging,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(opts)
		},
	}

	// TODO: Set variables for default options
	cmd.PersistentFlags().StringVar(
		&opts.logLevel,
		"log-level",
		"info",
		fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()),
	)

	cmd.PersistentFlags().StringVar(
		&opts.config,
		"config",
		"config.yaml",
		"the config file in yaml format",
	)

	cmd.PersistentFlags().BoolVar(
		&opts.confirm,
		"confirm",
		false,
		"false by default means that we do not push anything to google groups",
	)

	cmd.PersistentFlags().BoolVar(
		&opts.print,
		"print",
		false,
		"print the existing group information",
	)

	AddCommands(cmd)
	return cmd
}

func AddCommands(topLevel *cobra.Command) {
	// New commands can be added here.
	// addCommandName(topLevel)
}

func initLogging(*cobra.Command, []string) error {
	// TODO: Make this configurable
	return log.SetupGlobalLogger("info")
}
