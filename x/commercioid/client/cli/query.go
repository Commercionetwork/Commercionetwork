package cli

/**
This file contains the functions that returns the commands allowing the user to perform queries on the data stored
inside the blockchain.

Each query available should have an associated function here returning the proper command to execute from the CLI.

The query path should be contained inside the querier.go file too.
*/

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func GetCmdResolveIdentity(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "resolve [did]",
		Short: "Resolve identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			name := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/identities/%s", queryRoute, name), nil)
			if err != nil {
				fmt.Printf("Could not resolve identity - %s \n", string(name))
				return nil
			}

			fmt.Println(string(res))

			return nil
		},
	}
}

func GetCmdReadConnections(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "connections [did]",
		Short: "Lists all the connections associated to the given Did",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			name := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/connections/%s", queryRoute, name), nil)
			if err != nil {
				fmt.Printf("Could not get connections for %s: %s \n", string(name), err)
				return nil
			}

			fmt.Println(string(res))

			return nil
		},
	}
}