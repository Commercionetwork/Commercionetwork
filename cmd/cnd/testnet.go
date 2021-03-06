package main

// DONTCOVER

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/commercionetwork/commercionetwork/x/commerciokyc"
	commerciokycTypes "github.com/commercionetwork/commercionetwork/x/commerciokyc/types"
	governmentTypes "github.com/commercionetwork/commercionetwork/x/government/types"
	vbrTypes "github.com/commercionetwork/commercionetwork/x/vbr/types"

	"github.com/commercionetwork/commercionetwork/x/vbr"

	"github.com/commercionetwork/commercionetwork/x/government"

	"github.com/commercionetwork/commercionetwork/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNumValidators     = "v"
	flagOutputDir         = "output-dir"
	flagNodeDaemonHome    = "node-daemon-home"
	flagNodeCLIHome       = "node-cli-home"
	flagStartingIPAddress = "starting-ip-address"
)

// get cmd to initialize all files for tendermint testnet and application
func testnetCmd(ctx *server.Context, cdc *codec.Codec,
	mbm module.BasicManager, genBalIterator genutiltypes.GenesisAccountsIterator,
) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a Commercio testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	cnd testnet --v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := ctx.Config

			outputDir := viper.GetString(flagOutputDir)
			chainID := viper.GetString(flags.FlagChainID)
			minGasPrices := viper.GetString(server.FlagMinGasPrices)
			nodeDirPrefix := viper.GetString(flagNodeDirPrefix)
			nodeDaemonHome := viper.GetString(flagNodeDaemonHome)
			nodeCLIHome := viper.GetString(flagNodeCLIHome)
			startingIPAddress := viper.GetString(flagStartingIPAddress)
			numValidators := viper.GetInt(flagNumValidators)

			return InitTestnet(cmd, config, cdc, mbm, genBalIterator, outputDir, chainID,
				minGasPrices, nodeDirPrefix, nodeDaemonHome, nodeCLIHome, startingIPAddress, numValidators)
		},
	}

	cmd.Flags().Int(flagNumValidators, 4,
		"Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./mytestnet",
		"Directory to store initialization data for the testnet")
	cmd.Flags().String(flagNodeDirPrefix, "node",
		"Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "cnd",
		"Home directory of the node's daemon configuration")
	cmd.Flags().String(flagNodeCLIHome, "cncli",
		"Home directory of the node's cli configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1",
		"Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(
		flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(
		//server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", app.DefaultBondDenom),
		//"Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake)")
		server.FlagMinGasPrices, fmt.Sprintf(""),
		"No gas prices need setup")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")

	return cmd
}

const nodeDirPerm = 0755

// Initialize the testnet
func InitTestnet(
	cmd *cobra.Command, config *tmconfig.Config, cdc *codec.Codec,
	mbm module.BasicManager, genBalIterator genutiltypes.GenesisAccountsIterator,
	outputDir, chainID, minGasPrices, nodeDirPrefix, nodeDaemonHome,
	nodeCLIHome, startingIPAddress string, numValidators int,
) error {

	if chainID == "" {
		chainID = "chain-" + tmrand.NewRand().Str(6)
	}

	monikers := make([]string, numValidators)
	nodeIDs := make([]string, numValidators)
	valPubKeys := make([]crypto.PubKey, numValidators)

	cndConfig := srvconfig.DefaultConfig()
	cndConfig.MinGasPrices = minGasPrices

	//nolint:prealloc
	var (
		genAccounts []authexported.GenesisAccount
		genFiles    []string
	)

	inBuf := bufio.NewReader(cmd.InOrStdin())
	// generate private keys, node IDs, and initial transactions
	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		clientDir := filepath.Join(outputDir, nodeDirName, nodeCLIHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")

		config.SetRoot(nodeDir)
		config.RPC.ListenAddress = "tcp://0.0.0.0:26657"

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		if err := os.MkdirAll(clientDir, nodeDirPerm); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		monikers = append(monikers, nodeDirName)
		config.Moniker = nodeDirName

		ip, err := getIP(i, startingIPAddress)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(config)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		memo := fmt.Sprintf("%s@%s:26656", nodeIDs[i], ip)
		genFiles = append(genFiles, config.GenesisFile())

		kb, err := keys.NewKeyring(
			sdk.KeyringServiceName(),
			viper.GetString(flags.FlagKeyringBackend),
			clientDir,
			inBuf,
		)
		if err != nil {
			return err
		}

		keyPass := clientkeys.DefaultKeyPass
		addr, secret, err := server.GenerateSaveCoinKey(kb, nodeDirName, keyPass, true)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		info := map[string]string{"secret": secret}

		cliPrint, err := json.Marshal(info)
		if err != nil {
			return err
		}

		// save private key seed words
		if err := writeFile(fmt.Sprintf("%v.json", "key_seed"), clientDir, cliPrint); err != nil {
			return err
		}

		accTokens := sdk.TokensFromConsensusPower(10000000)
		accStakingTokens := sdk.TokensFromConsensusPower(10000000)
		coins := sdk.Coins{
			sdk.NewCoin(app.StableCreditsDenom, accTokens),
			sdk.NewCoin(app.DefaultBondDenom, accStakingTokens),
		}

		genAccounts = append(genAccounts, auth.NewBaseAccount(addr, coins.Sort(), nil, 0, 0))

		valTokens := sdk.TokensFromConsensusPower(100)
		msg := staking.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			valPubKeys[i],
			sdk.NewCoin(app.DefaultBondDenom, valTokens),
			staking.NewDescription(nodeDirName, "", "", "", ""),
			staking.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
			sdk.OneInt(),
		)

		tx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{}, memo)
		txBldr := auth.NewTxBuilderFromCLI(inBuf).WithChainID(chainID).WithMemo(memo).WithKeybase(kb)

		signedTx, err := txBldr.SignStdTx(nodeDirName, clientkeys.DefaultKeyPass, tx, false)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		txBytes, err := cdc.MarshalJSON(signedTx)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		// gather gentxs folder
		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBytes); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		// TODO: Rename config file to server.toml as it's not particular to Gaia
		// (REF: https://github.com/cosmos/cosmos-sdk/issues/4125).
		cndConfigFilePath := filepath.Join(nodeDir, "config/cnd.toml")
		srvconfig.WriteConfigFile(cndConfigFilePath, cndConfig)
	}

	if err := initGenFiles(cdc, mbm, chainID, genAccounts, genFiles, numValidators); err != nil {
		return err
	}

	err := collectGenFiles(
		cdc, config, chainID, monikers, nodeIDs, valPubKeys, numValidators,
		outputDir, nodeDirPrefix, nodeDaemonHome, genBalIterator,
	)
	if err != nil {
		return err
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", numValidators)
	return nil
}

func initGenFiles(
	cdc *codec.Codec, mbm module.BasicManager, chainID string,
	genAccounts []authexported.GenesisAccount,
	genFiles []string, numValidators int,
) error {

	appGenState := mbm.DefaultGenesis()

	// set the accounts in the genesis state
	var authGenState auth.GenesisState
	cdc.MustUnmarshalJSON(appGenState[auth.ModuleName], &authGenState)
	authGenState.Accounts = genAccounts
	appGenState[auth.ModuleName] = cdc.MustMarshalJSON(authGenState)

	// cnd set-genesis-government-address
	// cnd set-genesis-tumbler-address
	var governmentState government.GenesisState
	cdc.MustUnmarshalJSON(appGenState[governmentTypes.ModuleName], &authGenState)
	governmentState.GovernmentAddress = genAccounts[0].GetAddress()
	governmentState.TumblerAddress = genAccounts[0].GetAddress()
	appGenState[governmentTypes.ModuleName] = cdc.MustMarshalJSON(governmentState)

	// set-genesis-vbr-pool-amount 1000000000ucommercio
	var vbrState vbr.GenesisState
	cdc.MustUnmarshalJSON(appGenState[vbrTypes.ModuleName], &vbrState)
	tokens := sdk.TokensFromConsensusPower(1000)
	vbrState.PoolAmount = sdk.NewDecCoinsFromCoins(sdk.NewCoin(app.DefaultBondDenom, tokens))
	vbrState.AutomaticWithdraw = true
	vbrState.RewardRate = sdk.NewDecWithPrec(1, 3)
	appGenState[vbrTypes.ModuleName] = cdc.MustMarshalJSON(vbrState)

	// cnd add-genesis-tsp
	var kycState commerciokyc.GenesisState
	cdc.MustUnmarshalJSON(appGenState[commerciokycTypes.ModuleName], &kycState)
	kycState.TrustedServiceProviders, _ = kycState.TrustedServiceProviders.AppendIfMissing(genAccounts[0].GetAddress())
	// cnd add-genesis-membership black
	currentTime := time.Now()
	expiryAt, _ := time.Parse("2006-01-02", currentTime.Format("2006-01-02"))
	expiryAt = expiryAt.Add(time.Hour * 24 * 365)

	membership := commerciokycTypes.NewMembership("black", genAccounts[0].GetAddress(), genAccounts[0].GetAddress(), expiryAt)
	kycState.Memberships, _ = kycState.Memberships.AppendIfMissing(membership)
	genesisStateBz := cdc.MustMarshalJSON(kycState)
	appGenState[commerciokycTypes.ModuleName] = genesisStateBz

	appGenStateJSON, err := codec.MarshalJSONIndent(cdc, appGenState)
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}
	return nil
}

func collectGenFiles(
	cdc *codec.Codec, config *tmconfig.Config, chainID string,
	monikers, nodeIDs []string, valPubKeys []crypto.PubKey,
	numValidators int, outputDir, nodeDirPrefix, nodeDaemonHome string,
	genBalIterator genutiltypes.GenesisAccountsIterator,
) error {

	var appState json.RawMessage
	genTime := tmtime.Now()

	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")
		moniker := monikers[i]
		config.Moniker = nodeDirName

		config.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutil.NewInitConfig(chainID, gentxsDir, moniker, nodeID, valPubKey)

		genDoc, err := types.GenesisDocFromFile(config.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genutil.GenAppStateFromConfig(cdc, config, initCfg, *genDoc, genBalIterator)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := config.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)

	err := tmos.EnsureDir(writePath, 0700)
	if err != nil {
		return err
	}

	err = tmos.WriteFile(file, contents, 0600)
	if err != nil {
		return err
	}

	return nil
}
