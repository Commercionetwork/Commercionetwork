package accreditations

import (
	"errors"

	"github.com/commercionetwork/commercionetwork/x/accreditations/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - docs genesis state
type GenesisState struct {
	LiquidityPoolAmount sdk.Coins             `json:"liquidity_pool_amount"`
	Accreditations      []types.Accreditation `json:"users_data"`
	TrustworthySigners  []sdk.AccAddress      `json:"trustworthy_signers"`
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// InitGenesis sets docs information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	// Set the liquidity pool
	if err := keeper.DepositIntoPool(ctx, data.LiquidityPoolAmount); err != nil {
		panic(err)
	}

	// Import the signers
	for _, signer := range data.TrustworthySigners {
		keeper.AddTrustworthySigner(ctx, signer)
	}

	// Import all the accreditations
	for _, accreditation := range data.Accreditations {
		if err := keeper.SetAccrediter(ctx, accreditation.User, accreditation.Accrediter); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return GenesisState{
		LiquidityPoolAmount: keeper.GetPoolFunds(ctx),
		Accreditations:      keeper.GetAccreditations(ctx),
		TrustworthySigners:  keeper.GetTrustworthySigners(ctx),
	}
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if data.LiquidityPoolAmount.IsAnyNegative() {
		return errors.New("liquidity pool amount cannot contain negative values")
	}

	return nil
}
