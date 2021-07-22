package eak

import (
	"encoding/base64"

	adminAPI "github.com/smallstep/certificates/authority/admin/api"
	"github.com/smallstep/certificates/ca"
	"github.com/urfave/cli"
)

type cliEAK struct {
	id   string
	name string
	key  string
}

func toCLI(ctx *cli.Context, client *ca.AdminClient, eak *adminAPI.CreateExternalAccountKeyResponse) (*cliEAK, error) {
	// TODO: more fields for other purposes, like including the createdat/boundat/account for listing?
	return &cliEAK{id: eak.KeyID, name: eak.Name, key: base64.StdEncoding.EncodeToString(eak.Key)}, nil
}

// Command returns the jwk subcommand.
func Command() cli.Command {
	return cli.Command{
		Name:      "eak",
		Usage:     "create and manage ACME External Account Key material",
		UsageText: "**step beta ca eak** <subcommand> [arguments] [global-flags] [subcommand-flags]",
		Subcommands: cli.Commands{
			listCommand(),
			addCommand(),
			// removeCommand(),
			// updateCommand(),
		},
		Description: `**step beta ca eak** command group provides facilities for managing ACME 
		External Account Keys.

## EXAMPLES

List the active ACME External Account Keys:
'''
$ step beta ca eak list
'''

Add an ACME External Account Key:
'''
$ step beta ca eak add some_name_or_reference
'''`,
	}
}
