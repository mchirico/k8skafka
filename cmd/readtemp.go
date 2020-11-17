/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"context"
	"fmt"
	"github.com/mchirico/k8skafka/temperature/pubsub"
	"time"

	"github.com/spf13/cobra"
)

// readtempCmd represents the readtemp command
var readtempCmd = &cobra.Command{
	Use:   "readtemp",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("readtemp called. timeOuts 100 hard coded")
		timeOuts := 10000
		broker := "broker:9092"
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeOuts)*time.Second)
		defer cancel()

		pubsub.Sub(ctx,broker, timeOuts)
	},
}

func init() {
	rootCmd.AddCommand(readtempCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// readtempCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// readtempCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
