// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestIsRequiredFlag tests IsRequiredFlagNotFound and MarkFlagRequired
func TestIsRequiredFlag(t *testing.T) {
	convey.Convey("check whether the required field is missing", t, func() {
		originalCmdline := flag.CommandLine
		originalArgs := os.Args
		defer func() {
			flag.CommandLine = originalCmdline
			os.Args = originalArgs
		}()

		const flagName = "flag_var"

		var flagVar int
		flagSet1 := flag.NewFlagSet("", flag.ExitOnError)
		flagSet1.IntVar(&flagVar, flagName, 0, "usage")
		MarkFlagRequired(flagName)
		os.Args = []string{"test"}
		flag.CommandLine = flagSet1
		flag.Parse()
		convey.So(IsRequiredFlagNotFound(), convey.ShouldBeTrue)

		flagSet2 := flag.NewFlagSet("", flag.ExitOnError)
		flagSet2.IntVar(&flagVar, flagName, 0, "usage")
		MarkFlagRequired(flagName)
		os.Args = []string{"test", fmt.Sprintf("-%s=0", flagName)}
		flag.CommandLine = flagSet2
		flag.Parse()
		convey.So(IsRequiredFlagNotFound(), convey.ShouldBeFalse)

		flagSet3 := flag.NewFlagSet("", flag.ExitOnError)
		flagSet3.IntVar(&flagVar, flagName, 1, "usage")
		flag.CommandLine = flagSet3
		flag.Parse()
		convey.So(IsRequiredFlagNotFound(), convey.ShouldBeFalse)
	})
}

func TestIsFlagSet(t *testing.T) {
	convey.Convey("test IsFlagSet", t, func() {
		convey.So(IsFlagSet(""), convey.ShouldBeFalse)
		if err := flag.Set("testFlag", "abc"); err != nil {
			fmt.Printf("set test flag failed, error: %v\n", err)
			return
		}
		convey.So(IsFlagSet("testFlag"), convey.ShouldBeTrue)
	})
}
