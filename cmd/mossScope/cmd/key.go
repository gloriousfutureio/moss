// Copyright © 2017 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/couchbase/moss"
	"github.com/spf13/cobra"
)

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Dumps the key and value of the specified key",
	Long: `Dumps the key and value information of the requested key
from the latest snapshot in which it is available in JSON
format. For example:
	./mossScope dump key <keyname> <path_to_store> [flag]`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("USAGE: mossScope dump key <keyname> <path_to_store> " +
			            "[flag], more details with --help");
			return
		}

		key(args[0], args[1])
	},
}

var allVersions bool

func key(keyname string, dir string) {
	store, err := moss.OpenStore(dir, moss.StoreOptions{})
	if err != nil || store == nil {
		fmt.Printf("Moss-OpenStore() API failed, err: %v\n", err)
		os.Exit(-1)
	}

	snap, err := store.Snapshot()
	if err != nil || snap == nil {
		fmt.Printf("Store-Snapshot() API failed, err: %v\n", err)
		os.Exit(-1)
	}

	curr_snapshot := snap
	val, err := curr_snapshot.Get([]byte(keyname), moss.ReadOptions{})
	if (err == nil && val != nil) {
		fmt.Printf("[\n")
		dumpKeyVal([]byte(keyname), val, inHex)
	} else {
		snap.Close()
		store.Close()
		fmt.Printf("Key: '%s', does not exist!\n", keyname)
		return
	}

	if allVersions {
		prev_snapshot, err := store.SnapshotPrevious(curr_snapshot)
		for {
			if err != nil || prev_snapshot == nil {
				break
			}

			curr_snapshot = prev_snapshot
			val, err := curr_snapshot.Get([]byte(keyname), moss.ReadOptions{})
			if (err == nil && val != nil) {
				fmt.Printf(",\n")
				dumpKeyVal([]byte(keyname), val, inHex)
			}

			prev_snapshot, err = store.SnapshotPrevious(curr_snapshot)
			curr_snapshot.Close()
		}
	}
	fmt.Printf("\n]\n")

	snap.Close()
	store.Close()
}

// The following wrapper (public) is for test purposes
func Key(keyname string, dir string, getAllVersions bool) {
	allVersions = getAllVersions
	key(keyname, dir)
}

func init() {
	dumpCmd.AddCommand(keyCmd)

	// Local flag that is intended to work as a flag over dump key
	keyCmd.Flags().BoolVar(&allVersions, "all-versions", false,
	                       "Emits all the available versions of the key")
}
