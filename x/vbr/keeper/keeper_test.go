package keeper

import (
	"testing"

	sdkErr "github.com/cosmos/cosmos-sdk/types/errors"
	dist "github.com/cosmos/cosmos-sdk/x/distribution"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/commercionetwork/commercionetwork/x/vbr/types"
)

// -------------
// --- Pool
// -------------

func TestKeeper_SetTotalRewardPool(t *testing.T) {
	cdc, ctx, k, _, _, _ := SetupTestInput(true)

	k.SetTotalRewardPool(ctx, TestBlockRewardsPool)

	var pool sdk.DecCoins
	store := ctx.KVStore(k.storeKey)
	cdc.MustUnmarshalBinaryBare(store.Get([]byte(types.PoolStoreKey)), &pool)

	require.Equal(t, TestBlockRewardsPool, pool)
}

func TestKeeper_GetTotalRewardPool(t *testing.T) {
	tests := []struct {
		name         string
		pool         sdk.DecCoins
		expectedPool sdk.DecCoins
	}{
		{
			name:         "Get total from empty pool",
			pool:         sdk.DecCoins{sdk.NewInt64DecCoin("stake", 0)},
			expectedPool: sdk.DecCoins{},
		},
		{
			name:         "Get total from existing pool",
			pool:         sdk.DecCoins{sdk.NewInt64DecCoin("stake", 100)},
			expectedPool: sdk.DecCoins{sdk.NewInt64DecCoin("stake", 100)},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cdc, ctx, k, _, _, _ := SetupTestInput(true)
			store := ctx.KVStore(k.storeKey)
			store.Set([]byte(types.PoolStoreKey), cdc.MustMarshalBinaryBare(&tt.pool))

			macc := k.VbrAccount(ctx)
			suppl, _ := tt.pool.TruncateDecimal()
			_ = macc.SetCoins(sdk.NewCoins(suppl...))
			k.supplyKeeper.SetModuleAccount(ctx, macc)

			actual := k.GetTotalRewardPool(ctx)

			require.Equal(t, tt.expectedPool, actual)

		})
	}
}

// ---------------------------
// --- Reward distribution
// ---------------------------
func TestKeeper_ComputeProposerReward(t *testing.T) {
	tests := []struct {
		name           string
		bonded         sdk.Int
		vNumber        int64
		expectedReward string
	}{
		{
			"Compute reward with 100 validators",
			sdk.NewInt(100000000),
			100,
			"92.592592592592592593",
		},
		{
			"Compute reward with 50 validators",
			sdk.NewInt(100000000),
			50,
			"46.296296296296296296",
		},
		{
			"Compute reward with small bonded",
			sdk.NewInt(1),
			100,
			"0.000000925925925926",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, ctx, k, _, _, _ := SetupTestInput(false)

			testVal := TestValidator.UpdateStatus(sdk.Bonded)
			testVal, _ = testVal.AddTokensFromDel(tt.bonded)
			k.SetRewardRate(ctx, TestRewarRate)

			reward := k.ComputeProposerReward(ctx, tt.vNumber, testVal, "stake")

			expectedDecReward, _ := sdk.NewDecFromStr(tt.expectedReward)

			expected := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", expectedDecReward)}

			require.Equal(t, expected, reward)

		})
	}
}

func TestKeeper_DistributeBlockRewards(t *testing.T) {
	tests := []struct {
		name              string
		pool              sdk.DecCoins
		expectedValidator sdk.DecCoins
		expectedRemaining sdk.DecCoins
		expectedError     error
		bonded            sdk.Int
	}{
		{
			name:              "Reward with enough pool",
			pool:              sdk.DecCoins{sdk.NewInt64DecCoin("stake", 10000)},
			expectedRemaining: sdk.DecCoins{sdk.NewInt64DecCoin("stake", 9991)},
			expectedValidator: sdk.DecCoins{sdk.NewInt64DecCoin("stake", 9)},
			bonded:            sdk.NewInt(1000000000),
		},
		{
			name:              "Reward with empty pool",
			pool:              sdk.DecCoins{sdk.NewInt64DecCoin("stake", 0)},
			expectedRemaining: sdk.DecCoins{},
			expectedValidator: sdk.DecCoins(nil),
			bonded:            sdk.NewInt(1000000000),
		},
		{
			name:              "Reward not enough funds into pool",
			pool:              sdk.DecCoins{sdk.NewInt64DecCoin("stake", 1)},
			expectedRemaining: sdk.DecCoins{sdk.NewInt64DecCoin("stake", 1)},
			expectedError:     sdkErr.Wrap(sdkErr.ErrInsufficientFunds, "Pool hasn't got enough funds to supply validator's rewards"),
			expectedValidator: sdk.DecCoins(nil),
			bonded:            sdk.NewInt(1000000000),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, ctx, k, _, _, _ := SetupTestInput(false)

			testVal := TestValidator.UpdateStatus(sdk.Bonded)
			testVal, _ = testVal.AddTokensFromDel(tt.bonded)

			k.SetRewardRate(ctx, TestRewarRate)
			k.SetTotalRewardPool(ctx, tt.pool)

			macc := k.VbrAccount(ctx)
			suppl, _ := tt.pool.TruncateDecimal()
			_ = macc.SetCoins(sdk.NewCoins(suppl...))
			k.supplyKeeper.SetModuleAccount(ctx, macc)

			validatorRewards := dist.ValidatorCurrentRewards{Rewards: sdk.DecCoins{}}
			k.distKeeper.SetValidatorCurrentRewards(ctx, testVal.GetOperator(), validatorRewards)

			validatorOutstandingRewards := sdk.DecCoins{}
			k.distKeeper.SetValidatorOutstandingRewards(ctx, testVal.GetOperator(), validatorOutstandingRewards)

			reward := k.ComputeProposerReward(ctx, 1, testVal, "stake")

			err := k.DistributeBlockRewards(ctx, testVal, reward)
			if tt.expectedError != nil {
				require.Equal(t, err.Error(), tt.expectedError.Error())
			}

			valCurReward := k.distKeeper.GetValidatorCurrentRewards(ctx, testVal.GetOperator())
			//rewardedVal := valCurReward.Rewards
			rewardPool := k.GetTotalRewardPool(ctx)

			require.Equal(t, tt.expectedRemaining, rewardPool)
			require.Equal(t, tt.expectedValidator, valCurReward.Rewards)

		})
	}
}

/*
func TestKeeper_WithdrawAllRewards(t *testing.T) {
	tests := []struct {
		name            string
		bonded          sdk.Int
		bondedVal       sdk.Int
		rewardStr       string
		commisionStr    string
		expectedAccount sdk.Coins
		expectedVal     sdk.Coins
	}{
		{
			name:            "Reward",
			bonded:          sdk.NewInt(100000000),
			bondedVal:       sdk.NewInt(1000000),
			rewardStr:       "10.1",
			commisionStr:    "1",
			expectedAccount: sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			expectedVal:     sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, ctx, k, _, bk, sk := SetupTestInput(false)
			reward, _ := sdk.NewDecFromStr(tt.rewardStr)
			commision, _ := sdk.NewDecFromStr(tt.commisionStr)

			testVal := TestValidator.UpdateStatus(sdk.Bonded)
			testVal, _ = testVal.AddTokensFromDel(tt.bonded)

			bk.SetCoins(ctx, TestDelegator, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000000))))
			sk.SetValidator(ctx, testVal)

			delegationVal := stakingTypes.NewDelegation(sdk.AccAddress(testVal.GetOperator()), testVal.GetOperator(), sdk.NewDec(1000))
			sk.SetDelegation(ctx, delegationVal)
			delegation := stakingTypes.NewDelegation(TestDelegator, testVal.GetOperator(), sdk.NewDec(1000))
			sk.SetDelegation(ctx, delegation)

			validatorRewards := dist.ValidatorCurrentRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", reward))}
			k.distKeeper.SetValidatorCurrentRewards(ctx, testVal.GetOperator(), validatorRewards)

			validatorOutstandingRewards := sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", reward))
			k.distKeeper.SetValidatorOutstandingRewards(ctx, testVal.GetOperator(), validatorOutstandingRewards)

			k.distKeeper.SetValidatorAccumulatedCommission(ctx, testVal.GetOperator(), sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", commision)))

			k.WithdrawAllRewards(ctx, sk)
			valCurReward := k.distKeeper.GetValidatorCurrentRewards(ctx, testVal.GetOperator())
			fmt.Println(valCurReward)
			//actualDel := bk.GetCoins(ctx, TestDelegator)
			actualVal := bk.GetCoins(ctx, sdk.AccAddress(testVal.GetOperator()))

			//require.Equal(t, tt.expectedAccount, actualDel)
			require.Equal(t, tt.expectedVal, actualVal)

		})
	}
}*/

func TestKeeper_VbrAccount(t *testing.T) {
	tests := []struct {
		name              string
		wantModName       string
		wantModAccBalance sdk.Coins
		emptyPool         bool
	}{
		{
			"an empty vbr account",
			"vbr",
			sdk.NewCoins(),
			true,
		},
		{
			"a vbr account with coins in it",
			"vbr",
			sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(100000), Denom: "stake"}),
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, ctx, k, _, _, _ := SetupTestInput(tt.emptyPool)
			macc := k.VbrAccount(ctx)

			require.Equal(t, macc.GetName(), tt.wantModName)
			require.True(t, macc.GetCoins().IsEqual(tt.wantModAccBalance))
		})
	}
}

func TestKeeper_MintVBRTokens(t *testing.T) {
	tests := []struct {
		name       string
		wantAmount sdk.Coins
	}{
		{
			"add 10stake",
			sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(10), Denom: "stake"}),
		},
		{
			"add no stake",
			sdk.NewCoins(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, ctx, k, _, _, _ := SetupTestInput(true)
			k.MintVBRTokens(ctx, tt.wantAmount)
			macc := k.VbrAccount(ctx)
			require.True(t, macc.GetCoins().IsEqual(tt.wantAmount))
		})
	}
}
