package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var calibrateCmd = &cobra.Command{
	Use:   "calibrate",
	Short: "Calibrate plan steps sequentially",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not implemented: calibrate")
	},
}

func init() {
	f := calibrateCmd.Flags()
	f.String("plan", "", "path to .plan/ directory")
	f.Int("runs", 3, "number of calibration runs per step")
	f.Bool("keep-testbed", false, "keep testbed after calibration")
	f.Int("max-retries", 2, "max patch retries per failing run")
	f.String("model", "", "claude model for step execution")
	f.String("patch-model", "", "claude model for patch generation")
	f.Int("concurrency", 1, "parallel calibration runs")
	f.Duration("timeout", 30*time.Minute, "timeout per step run")
	f.Bool("dry", false, "dry run: validate plan without execution")

	rootCmd.AddCommand(calibrateCmd)
}
