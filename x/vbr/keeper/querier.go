package keeper

import (
	"fmt"

	sdkErr "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	vbrTypes "github.com/commercionetwork/commercionetwork/x/vbr/types"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case vbrTypes.QueryBlockRewardsPoolFunds:
			return queryGetBlockRewardsPoolFunds(ctx, path[1:], keeper)
		case vbrTypes.QueryRewardRate:
			return queryRewardRate(ctx, keeper)

		case vbrTypes.QueryAutomaticWithdraw:
			return queryAutomaticWithdraw(ctx, keeper)

		default:
			return nil, sdkErr.Wrap(sdkErr.ErrUnknownRequest, fmt.Sprintf("Unknown %s query endpoint", vbrTypes.ModuleName))
		}
	}
}

func queryGetBlockRewardsPoolFunds(ctx sdk.Context, _ []string, keeper Keeper) (res []byte, err error) {
	funds := keeper.GetTotalRewardPool(ctx)

	fundsBz, err2 := codec.MarshalJSONIndent(keeper.cdc, funds)
	if err2 != nil {
		return nil, sdkErr.Wrap(sdkErr.ErrUnknownRequest, "Could not marshal result to JSON")
	}

	return fundsBz, nil
}

func queryRewardRate(ctx sdk.Context, keeper Keeper) ([]byte, error) {
	return codec.MarshalJSONIndent(keeper.cdc, keeper.GetRewardRate(ctx))
}

func queryAutomaticWithdraw(ctx sdk.Context, keeper Keeper) ([]byte, error) {
	return codec.MarshalJSONIndent(keeper.cdc, keeper.GetAutomaticWithdraw(ctx))
}
