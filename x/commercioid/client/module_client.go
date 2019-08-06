package client

import (
	commercioidcmd "commercio-network/x/commercioid/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:   "commercioid",
		Short: "CommercioID querying commands",
	}

	govQueryCmd.AddCommand(client.GetCommands(
		commercioidcmd.GetCmdResolveIdentity(mc.storeKey, mc.cdc),
		commercioidcmd.GetCmdReadConnections(mc.storeKey, mc.cdc),
	)...)

	return govQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:   "commercioid",
		Short: "CommercioID transactions subcommands",
	}

	govTxCmd.AddCommand(client.PostCommands(
		commercioidcmd.GetCmdSetIdentity(mc.cdc),
		commercioidcmd.GetCmdCreateConnection(mc.cdc),
	)...)

	return govTxCmd
}