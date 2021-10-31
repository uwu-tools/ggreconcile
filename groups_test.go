/*
Copyright 2019 The Kubernetes Authors.

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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"k8s.io/apimachinery/pkg/util/sets"
)

var cfg GroupsConfig
var rConfig RestrictionsConfig

var groupsPath = flag.String("groups-path", "", "Directory containing groups.yaml files")
var restrictionsPath = flag.String("restrictions-path", "", "Path to the configuration file containing restrictions")

func TestMain(m *testing.M) {
	flag.Parse()
	var err error

	if *restrictionsPath != "" && !filepath.IsAbs(*restrictionsPath) {
		fmt.Printf("restrictions-path \"%s\" must be an absolute path\n", *restrictionsPath)
		os.Exit(1)
	}

	if *restrictionsPath == "" {
		baseDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Cannot get current working directory: %v\n", err)
			os.Exit(1)
		}
		rPath := filepath.Join(baseDir, defaultRestrictionsFile)
		restrictionsPath = &rPath
	}

	if err := rConfig.Load(*restrictionsPath); err != nil {
		fmt.Printf("Could not load restrictions config: %v\n", err)
		os.Exit(1)
	}

	if *groupsPath != "" && !filepath.IsAbs(*groupsPath) {
		fmt.Printf("groups-path \"%s\" must be an absolute path\n", *groupsPath)
		os.Exit(1)
	}

	if *groupsPath == "" {
		*groupsPath, err = os.Getwd()
		if err != nil {
			fmt.Printf("Cannot get current working directory: %v\n", err)
			os.Exit(1)
		}
	}

	if err := cfg.Load(*groupsPath, &rConfig); err != nil {
		fmt.Printf("Could not load groups config: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// TestMergedGroupsConfig tests that readGroupsConfig reads all
// groups.yaml files and the merged config does not contain any duplicates.
//
// It tests that the config is merged by checking that the final
// GroupsConfig contains at least one group that isn't in the
// root groups.yaml file.
func TestMergedGroupsConfig(t *testing.T) {
	var containsMergedConfig bool
	found := sets.String{}
	dups := sets.String{}

	for _, g := range cfg.Groups {
		name := g.Name
		if name == "ws-oss" {
			containsMergedConfig = true
		}

		if found.Has(name) {
			dups.Insert(name)
		}
		found.Insert(name)
	}

	if !containsMergedConfig {
		t.Errorf("Final GroupsConfig does not have merged configs from all groups.yaml files")
	}
	if n := len(dups); n > 0 {
		t.Errorf("%d duplicate groups: %s", n, strings.Join(dups.List(), ", "))
	}
}

// TestDescriptionLength tests that the number of characters in the
// google groups description does not exceed 300.
//
// This validation is needed because gcloud allows apps:description
// with length no greater than 300
func TestDescriptionLength(t *testing.T) {
	var errs []error
	for _, g := range cfg.Groups {
		description := g.Description

		len := utf8.RuneCountInString(description)
		//Ref: https://developers.google.com/admin-sdk/groups-settings/v1/reference/groups
		if len > 300 {
			errs = append(errs,
				fmt.Errorf("Number of characters in description \"%s\" for group name \"%s\" "+
					"should not exceed 300; is: %d", description, g.Name, len))
		}
	}

	if len(errs) > 0 {
		for _, err := range errs {
			t.Error(err)
		}
	}
}

// Enforce conventions for all groups
func TestGroupConventions(t *testing.T) {
	for _, g := range cfg.Groups {
		// groups are easier to reason about if email and name match
		expectedEmailId := g.Name + "@inclusivenaming.org"
		if g.EmailId != expectedEmailId {
			t.Errorf("group '%s': expected email '%s', got '%s'", g.Name, expectedEmailId, g.EmailId)
		}
	}
}

// TODO: Consider adding RBAC groups and accompanying tests
// TODO: Consider adding security response groups and accompanying tests

// An e-mail address can only show up once within a given group, whether that
// be as a member, manager, or owner
func TestNoDuplicateMembers(t *testing.T) {
	for _, g := range cfg.Groups {
		members := map[string]bool{}
		for _, m := range g.Members {
			if _, ok := members[m]; ok {
				t.Errorf("group '%s' cannot have duplicate member '%s'", g.EmailId, m)
			}
			members[m] = true
		}
		managers := map[string]bool{}
		for _, m := range g.Managers {
			if _, ok := members[m]; ok {
				t.Errorf("group '%s' manager '%s' cannot also be listed as a member", g.EmailId, m)
			}
			if _, ok := managers[m]; ok {
				t.Errorf("group '%s' cannot have duplicate manager '%s'", g.EmailId, m)
			}
			managers[m] = true
		}
		owners := map[string]bool{}
		for _, m := range g.Owners {
			if _, ok := members[m]; ok {
				t.Errorf("group '%s' owner '%s' cannot also be listed as a member", g.EmailId, m)
			}
			if _, ok := managers[m]; ok {
				t.Errorf("group '%s' owner '%s' cannot also be listed as a manager", g.EmailId, m)
			}
			if _, ok := owners[m]; ok {
				t.Errorf("group '%s' cannot have duplicate owner '%s'", g.EmailId, m)
			}
			owners[m] = true
		}
	}
}

// NOTE: make very certain you know what you are doing if you change one
// of these groups, we don't want to accidentally lock ourselves out
func TestHardcodedGroupsForParanoia(t *testing.T) {
	groups := map[string][]string{
		"billing@inclusivenaming.org": {
			"hagbard@gmail.com",
			"hey@auggie.dev",
		},
		"infra-admins@inclusivenaming.org": {
			"hagbard@gmail.com",
			"hey@auggie.dev",
		},
	}

	found := make(map[string]bool)

	for _, g := range cfg.Groups {
		if expected, ok := groups[g.EmailId]; ok {
			found[g.EmailId] = true
			sort.Strings(expected)
			actual := make([]string, len(g.Managers))
			copy(actual, g.Managers)
			sort.Strings(actual)
			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("group '%s': expected managers '%v', got '%v'", g.Name, expected, actual)
			}
		}
	}

	for email := range groups {
		if _, ok := found[email]; !ok {
			t.Errorf("group '%s' is missing, should be present", email)
		}
	}
}

// Setting AllowWebPosting should be set for every group which should support
// access to the group not only via gmail but also via web (you can see the list
// and history of threads and also use web interface to operate the group)
// More info:
// 	https://developers.google.com/admin-sdk/groups-settings/v1/reference/groups#allowWebPosting
func TestGroupsWhichShouldSupportHistory(t *testing.T) {
	groups := map[string]struct{}{
		"billing@inclusivenaming.org":      {},
		"infra-admins@inclusivenaming.org": {},
		"leads@inclusivenaming.org":        {},
	}

	found := make(map[string]struct{})

	for _, group := range cfg.Groups {
		emailId := group.EmailId
		found[emailId] = struct{}{}
		if _, ok := groups[emailId]; ok {
			allowedWebPosting, ok := group.Settings["AllowWebPosting"]
			if !ok {
				t.Errorf(
					"group '%s': must have 'settings.allowedWebPosting = true'",
					group.Name,
				)
			} else if allowedWebPosting != "true" {
				t.Errorf(
					"group '%s': must have 'settings.allowedWebPosting = true'"+
						" but have 'settings.allowedWebPosting = %s' instead",
					group.Name,
					allowedWebPosting,
				)
			}
		}
	}

	for email := range groups {
		if _, ok := found[email]; !ok {
			t.Errorf("group '%s' is missing, should be present", email)
		}
	}
}
