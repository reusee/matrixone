//go:build server_test
// +build server_test

package main

import (
	"testing"
)

func TestMOServer(t *testing.T) {
	NewScope().Fork(
		func() Arguments {
			return Arguments{
				"./system_vars_config.toml",
			}
		},
	).Call(func(
		main Main,
	) {
		main()
	})
}
