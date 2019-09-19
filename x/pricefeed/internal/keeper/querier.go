package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/commercionetwork/commercionetwork/x/pricefeed/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryGetTokenPrice:
			return queryGetTokenPrice(ctx, path[1:], keeper)
		case types.QueryGetOracles:
			return queryGetOracles(ctx, path[1:], keeper)
		default:
			return nil, sdk.ErrUnknownRequest(fmt.Sprintf("Unknown %s query endpoint", types.ModuleName))
		}
	}
}

func queryGetTokenPrice(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	tokenName := path[0]
	price := keeper.GetPrice(ctx, tokenName)
	priceBz, err := codec.MarshalJSONIndent(keeper.cdc, price)
	if err != nil {
		return nil, sdk.ErrUnknownRequest("Could not marshal result to JSON")
	}

	return priceBz, nil
}

func queryGetOracles(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {

	oracles := keeper.GetOracles(ctx)
	oraclesBz, err := codec.MarshalJSONIndent(keeper.cdc, oracles)
	if err != nil {
		return nil, sdk.ErrUnknownRequest("Could not marshal result to JSON")
	}

	return oraclesBz, nil
}
