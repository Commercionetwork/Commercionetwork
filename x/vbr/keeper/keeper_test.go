package keeper

import (
	//"fmt"

	"testing"

	sdkErr "github.com/cosmos/cosmos-sdk/types/errors"
	dist "github.com/cosmos/cosmos-sdk/x/distribution"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/commercionetwork/commercionetwork/x/vbr/types"
	//"github.com/cosmos/cosmos-sdk/x/staking"
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


/*func TestKeeper_WithdrawAllRewards(t *testing.T) {
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

			bk.SetCoins(ctx, valDelAddr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000000))))
			bk.SetCoins(ctx, sdk.AccAddress(valAddrVal), sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000000))))

			sk.SetValidator(ctx, testVal)

			val, found := sk.GetValidator(ctx, valAddrVal)
			if found {
				fmt.Println(val)
			}
			fmt.Println("----- ==================================== -----")

			//delegationVal := stakingTypes.NewDelegation(sdk.AccAddress(valAddrVal), valAddrVal, sdk.NewDec(1000))

			fmt.Println("----- delegation info xxx-----")
			delegationVal := staking.NewDelegation(sdk.AccAddress(valAddrVal), valAddrVal, sdk.NewDec(1000))
			sk.SetDelegation(ctx, delegationVal)
			fmt.Println(delegationVal)

			//func (k Keeper) SetDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress, period types.DelegatorStartingInfo) {

			//dist.Keeper.SetDelegatorStartingInfo(ctx, valAddrVal, dist.NewDelegatorStartingInfo(0, sdk.NewDec(0), 1))

			delegation := staking.NewDelegation(valDelAddr, valAddrVal, sdk.NewDec(1000))
			sk.SetDelegation(ctx, delegation)
			fmt.Println(delegation)

			//_, _ = sk.Delegate(ctx, valDelAddr, sdk.NewInt(100), sdk.Unbonded, testVal, true)

			validatorRewards := dist.ValidatorCurrentRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", reward))}
			k.distKeeper.SetValidatorCurrentRewards(ctx, valAddrVal, validatorRewards)

			previousPeriod := k.distKeeper.GetValidatorCurrentRewards(ctx, valAddrVal).Period - 1

			stake := testVal.TokensFromSharesTruncated(delegationVal.GetShares())
			k.distKeeper.SetDelegatorStartingInfo(ctx, valAddr, sdk.AccAddress(valAddrVal), dist.NewDelegatorStartingInfo(previousPeriod, stake, uint64(ctx.BlockHeight())))

			validatorOutstandingRewards := sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", reward))
			k.distKeeper.SetValidatorOutstandingRewards(ctx, valAddrVal, validatorOutstandingRewards)

			k.distKeeper.SetValidatorAccumulatedCommission(ctx, valAddrVal, sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", commision)))

			k.WithdrawAllRewards(ctx, sk)
			valCurReward := k.distKeeper.GetValidatorCurrentRewards(ctx, valAddrVal)
			fmt.Println(valCurReward)
			//actualDel := bk.GetCoins(ctx, TestDelegator)
			actualVal := bk.GetCoins(ctx, sdk.AccAddress(valAddrVal))

			//require.Equal(t, tt.expectedAccount, actualDel)
			require.Equal(t, tt.expectedVal, actualVal)

		})
	}
}
*/

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

//NOT WORKING
//panic --> GetValidatorHistoricalRewards not correctly inizializzed
/*func TestKeeper_WithdrawAllRewards(t *testing.T){
	tests := []struct {
		name            string
		bonded          sdk.Int
		rewardStr       string
		commisionStr    string
	}{
		{
			name:            "Reward",
			bonded:          sdk.NewInt(100000000),
			rewardStr:       "10.1",
			commisionStr:    "1",
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

			bk.SetCoins(ctx, valDelAddr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000000))))
			bk.SetCoins(ctx, sdk.AccAddress(valAddrVal), sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000000))))

			sk.SetValidator(ctx, testVal)

			_, found := sk.GetValidator(ctx, valAddrVal)
			
			if found {
				k.distKeeper.SetDelegatorStartingInfo(ctx, valAddrVal, TestDelegator, dist.NewDelegatorStartingInfo(0, sdk.NewDec(0), 1))
				//new delegation
				delegationVal := staking.NewDelegation(TestDelegator, valAddrVal, sdk.NewDec(1000))
				sk.SetDelegation(ctx, delegationVal)
				fmt.Println(delegationVal)
				
				//set validator's rewards
				validatorRewards := dist.ValidatorCurrentRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", reward))}
				
				//trying to inizialize the validatorHistoricalReward
				//ReferenceCount?
				valHistrewards := dist.NewValidatorHistoricalRewards(validatorRewards.Rewards, uint16(1))
				dist.Keeper.SetValidatorHistoricalRewards(k.distKeeper, ctx, valAddrVal, validatorRewards.Period, valHistrewards)
				
				k.distKeeper.SetValidatorCurrentRewards(ctx, valAddrVal, validatorRewards)

				//outstandingRewards 
				validatorOutstandingRewards := sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", reward))
				k.distKeeper.SetValidatorOutstandingRewards(ctx, valAddrVal, validatorOutstandingRewards)
				
				//set comissions
				k.distKeeper.SetValidatorAccumulatedCommission(ctx, valAddrVal, sdk.NewDecCoins(sdk.NewDecCoinFromDec("stake", commision)))
				//delegation rewards
				dist.NewDelegationDelegatorReward(valAddrVal, k.distKeeper.GetValidatorCurrentRewards(ctx, valAddrVal).Rewards)
				
				//execute the keeper method to test
				err := k.WithdrawAllRewards(ctx,sk)

				require.Nil(t, err)
			}
		})

	}
}
*/

func TestKeeper_GetRewardRate(t *testing.T){
	_, ctx, k, _, _, _ := SetupTestInput(true)
	store := ctx.KVStore(k.storeKey)

	store.Set([]byte(types.RewardRateKey), k.cdc.MustMarshalBinaryBare(TestRewarRate))

	actual := k.GetRewardRate(ctx)
	
	require.Equal(t, TestRewarRate, actual)
}

func TestKeeper_SetRewardRate(t *testing.T){
	_, ctx, k, _, _, _ := SetupTestInput(true)
	store := ctx.KVStore(k.storeKey)

	k.SetRewardRate(ctx, TestRewarRate)

	var actual sdk.Dec
	k.cdc.MustUnmarshalBinaryBare(store.Get([]byte(types.RewardRateKey)), &actual)

	require.Equal(t, TestRewarRate, actual)
}

func TestKeeper_GetAutomaticWithdraw(t *testing.T){
	_, ctx, k, _, _, _ := SetupTestInput(true)

	var autoWithdraw bool
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.AutomaticWithdraw), k.cdc.MustMarshalBinaryBare(autoWithdraw))

	actual := k.GetAutomaticWithdraw(ctx)

	require.Equal(t, autoWithdraw, actual)
}

func TestKeeper_SetAutomaticWithdraw(t *testing.T){
	_, ctx, k, _, _, _ := SetupTestInput(true)

	var autoWithdraw bool

	error := k.SetAutomaticWithdraw(ctx, autoWithdraw)
	
	require.Nil(t, error)

	store := ctx.KVStore(k.storeKey)
	var actual bool
	k.cdc.MustUnmarshalBinaryBare(store.Get([]byte(types.AutomaticWithdraw)), &actual)

	require.Equal(t, autoWithdraw, actual)
}

func TestKeeper_IsDailyWithdrawBlock(t *testing.T){
	_, _, k, _, _, _ := SetupTestInput(true)

	var lowHeight int64 = BPD.Int64() - 1
	lH := k.IsDailyWithdrawBlock(lowHeight)
	require.False(t, lH)

	var incorrectHeight int64 = BPD.Int64() + 1
	iH := k.IsDailyWithdrawBlock(incorrectHeight)
	require.False(t, iH)

	var rightHeight int64 = BPD.Int64()
	rH := k.IsDailyWithdrawBlock(rightHeight)
	require.True(t, rH)

	var zeroHeight int64 = 0
	zH := k.IsDailyWithdrawBlock(zeroHeight)
	require.False(t, zH)
	
}