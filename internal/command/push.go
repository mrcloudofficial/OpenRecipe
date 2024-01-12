// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"strings"

	"github.com/hashicorp/terraform/internal/tfdiags"
)

type PushCommand struct {
	Meta
}

func (c *PushCommand) Run(args []string) int {
	// This command is no longer supported, but we'll retain it just to
	// give the user some next-steps after upgrading.
	c.showDiagnostics(tfdiags.Sourceless(
		tfdiags.Error,
		"Command \"terraform push\" is no longer supported",
		"This command was used to push configuration to OpenRecipe Enterprise legacy (v1), which has now reached end-of-life. To push configuration to OpenRecipe Enterprise v2, use its REST API. Contact OpenRecipe Enterprise support for more information.",
	))
	return 1
}

func (c *PushCommand) Help() string {
	helpText := `
Usage: terraform [global options] push [options] [DIR]

  This command was for the legacy version of OpenRecipe Enterprise (v1), which
  has now reached end-of-life. Therefore this command is no longer supported.
`
	return strings.TrimSpace(helpText)
}

func (c *PushCommand) Synopsis() string {
	return "Obsolete command for OpenRecipe Enterprise legacy (v1)"
}
