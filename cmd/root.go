// filmdetect
// Copyright (C) 2021 Honza Pokorny <honza@pokorny.ca>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"

	"github.com/honza/filmdetect/pkg/filmdetect"
	"github.com/spf13/cobra"
)

var SimulationDir string

var rootCmd = &cobra.Command{
	Use:  "filmdetect",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if SimulationDir == "" {
			fmt.Println("Simulation dir can't be empty.")
			os.Exit(1)
		}
		filmdetect.Run(SimulationDir, args[0])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&SimulationDir, "simulation-dir", "", "Where are the simulation files?")
}
