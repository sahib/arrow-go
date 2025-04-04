// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !noasm
// +build !noasm

package bmi

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/klauspost/cpuid/v2"
)

var hasASIMD bool

func init() {
	// Added ability to enable extension via environment:
	// ARM_ENABLE_EXT=NEON go test
	if ext, ok := os.LookupEnv("ARM_ENABLE_EXT"); ok {
		if ext == "DISABLE" {
			cpuid.CPU.Disable(cpuid.ASIMD, cpuid.AESARM, cpuid.PMULL)
		} else {
			exts := strings.Split(ext, ",")

			for _, x := range exts {
				switch x {
				case "NEON":
					cpuid.CPU.Enable(cpuid.ASIMD)
				case "AES":
					cpuid.CPU.Enable(cpuid.AESARM)
				case "PMULL":
					cpuid.CPU.Enable(cpuid.PMULL)
				default:
					fmt.Fprintln(os.Stderr, "unrecognized value for ARM_ENABLE_EXT:", x)
				}
			}
		}
	}

	hasASIMD = cpuid.CPU.Has(cpuid.ASIMD)
}

//go:noescape
func _levels_to_bitmap_neon(levels unsafe.Pointer, numLevels int, rhs int16) (res uint64)

// greaterThanBitmapNEON builds a bitmap where each set bit indicates the corresponding level
// is greater than the rhs value.
func GreaterThanBitmap(levels []int16, rhs int16) uint64 {
	if !hasASIMD {
		return greaterThanBitmapGo(levels, rhs)
	}

	if len(levels) == 0 {
		return 0
	}

	var (
		p1 = unsafe.Pointer(&levels[0])
		p2 = len(levels)
		p3 = rhs
	)

	return _levels_to_bitmap_neon(p1, p2, p3)
}

func ExtractBits(bitmap, selectBitmap uint64) uint64 {
	return extractBitsGo(bitmap, selectBitmap)
}
