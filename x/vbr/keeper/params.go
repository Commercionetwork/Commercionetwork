package keeper

import (
	"github.com/commercionetwork/commercionetwork/x/vbr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParamSet returns the total set params
func (k Keeper) GetParamSet(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParamSet sets the total set of params
func (k Keeper) SetParamSet(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	k.paramSpace.SetParamSet(ctx, &params)
	return nil
}
