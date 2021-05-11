package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/commercionetwork/commercionetwork/x/vbr/types"
)

var request abci.RequestQuery

func TestQuerier_getBlockRewardsPoolFunds(t *testing.T) {
	var cdc, ctx, k, _, _, _ = SetupTestInput(false)
	var querier = NewQuerier(k)

	k.SetTotalRewardPool(ctx, TestBlockRewardsPool)

	path := []string{types.QueryBlockRewardsPoolFunds}
	actual, _ := querier(ctx, path, request)

	expected, _ := codec.MarshalJSONIndent(cdc, &TestBlockRewardsPool)

	require.Equal(t, expected, actual)
}

func TestQuerier_queryRewardRate(t *testing.T) {
	_, ctx, k, _, _, _ := SetupTestInput(false)
	require.NoError(t, k.SetRewardRate(ctx, sdk.NewDec(2)))
	querier := NewQuerier(k)
	actualBz, err := querier(ctx, []string{types.QueryRewardRate}, request)
	require.Nil(t, err)

	var rate sdk.Dec
	k.cdc.MustUnmarshalJSON(actualBz, &rate)
	require.Equal(t, sdk.NewDec(2), rate)
}

func TestQuerier_queryAutomaticWithdraw(t *testing.T) {
	_, ctx, k, _, _, _ := SetupTestInput(false)
	require.NoError(t, k.SetAutomaticWithdraw(ctx, true))
	querier := NewQuerier(k)
	actualBz, err := querier(ctx, []string{types.QueryAutomaticWithdraw}, request)
	require.Nil(t, err)

	var automaticWithdraw bool
	k.cdc.MustUnmarshalJSON(actualBz, &automaticWithdraw)
	require.Equal(t, true, automaticWithdraw)
}
