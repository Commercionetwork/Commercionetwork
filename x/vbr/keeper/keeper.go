package keeper

import (
	"fmt"

	sdkErr "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
	supplyExported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	ctypes "github.com/commercionetwork/commercionetwork/x/common/types"
	"github.com/commercionetwork/commercionetwork/x/vbr/types"
)

type Keeper struct {
	cdc          *codec.Codec
	storeKey     sdk.StoreKey
	distKeeper   distribution.Keeper
	supplyKeeper supply.Keeper
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, dk distribution.Keeper, sk supply.Keeper) Keeper {
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		distKeeper:   dk,
		supplyKeeper: sk,
	}
}

// -------------
// --- Pool
// -------------

// SetTotalRewardPool allows to set the value of the total rewards pool that has left
func (k Keeper) SetTotalRewardPool(ctx sdk.Context, updatedPool sdk.DecCoins) {
	store := ctx.KVStore(k.storeKey)
	if !updatedPool.Empty() {
		store.Set([]byte(types.PoolStoreKey), k.cdc.MustMarshalBinaryBare(&updatedPool))
	} else {
		store.Delete([]byte(types.PoolStoreKey))
	}
}

// GetTotalRewardPool returns the current total rewards pool amount
func (k Keeper) GetTotalRewardPool(ctx sdk.Context) sdk.DecCoins {
	macc := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
	mcoins := macc.GetCoins()

	return sdk.NewDecCoinsFromCoins(mcoins...)
}

// --------------------
// --- Year number
// --------------------

var (
	// DPY is Days Per Year
	DPY = sdk.NewDecWithPrec(36525, 2)
	// HPD is Hours Per Day
	HPD = sdk.NewDecWithPrec(24, 0)
	// MPH  is Minutes Per Hour
	MPH = sdk.NewDecWithPrec(60, 0)
	// BPM is Blocks Per Minutes
	BPM = sdk.NewDecWithPrec(9, 0)

	// BPD is Blocks Per Day
	BPD = HPD.Mul(MPH).Mul(BPM)

	// BPY is Blocks Per Year
	BPY = DPY.Mul(BPD)
)

// ---------------------------
// --- Reward distribution
// ---------------------------

// ComputeProposerReward computes the final reward for the validator block's proposer
func (k Keeper) ComputeProposerReward(ctx sdk.Context, validatorsCount int64,
	proposer exported.ValidatorI, totalStakedTokens sdk.Int) sdk.DecCoins {

	// Get rewarded rate
	rewardRate := k.GetRewardRate(ctx)

	// Calculate rewarded rate with validator percentage
	rewardRateVal := rewardRate.Mul(sdk.NewDec(validatorsCount)).Quo(sdk.NewDec(100))

	// Get total bonded token of validator
	proposerBonded := proposer.GetBondedTokens()

	// Compute reward for each block
	Rnb := sdk.NewDecCoinsFromCoins(sdk.NewCoin("ucommercio", proposerBonded.ToDec().Mul(rewardRateVal).Quo(BPD).TruncateInt()))

	return Rnb
}

// DistributeBlockRewards distributes the computed reward to the block proposer
func (k Keeper) DistributeBlockRewards(ctx sdk.Context, validator exported.ValidatorI, reward sdk.DecCoins) error {
	rewardPool := k.GetTotalRewardPool(ctx)
	// Check if the yearly pool and the total pool have enough funds
	if ctypes.IsAllGTE(rewardPool, reward) {
		// truncate fractional part and only take the integer part into account
		rewardInt, _ := reward.TruncateDecimal()

		k.SetTotalRewardPool(ctx, rewardPool.Sub(sdk.NewDecCoinsFromCoins(rewardInt...)))

		err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, distribution.ModuleName, rewardInt)
		if err != nil {
			return nil
		}

		k.distKeeper.AllocateTokensToValidator(ctx, validator, sdk.NewDecCoinsFromCoins(rewardInt...))
	} else {
		return sdkErr.Wrap(sdkErr.ErrInsufficientFunds, "Pool hasn't got enough funds to supply validator's rewards")
	}

	return nil
}

// WithdrawRewards withdraw reward to all validator
func (k Keeper) WithdrawAllRewards(ctx sdk.Context, stakeKeeper staking.Keeper) error {
	// Get all validators
	// Loop throw validators and withdraw all delegator rewards
	dels := stakeKeeper.GetAllDelegations(ctx)
	for _, delegation := range dels {
		_, _ = k.distKeeper.WithdrawDelegationRewards(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)
		k.distKeeper.WithdrawValidatorCommission(ctx, delegation.ValidatorAddress)
	}

	vals := stakeKeeper.GetAllValidators(ctx)
	for _, validator := range vals {
		k.distKeeper.WithdrawValidatorCommission(ctx, validator.GetOperator())
	}

	return nil
}

// VbrAccount returns vbr's ModuleAccount
func (k Keeper) VbrAccount(ctx sdk.Context) supplyExported.ModuleAccountI {
	return k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// MintVBRTokens mints coins into the vbr's ModuleAccount
func (k Keeper) MintVBRTokens(ctx sdk.Context, coins sdk.Coins) error {
	if err := k.supplyKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("could not mint requested coins: %w", err)
	}

	return nil
}

// GetRewardRate retrieve the vbr reward rate.
func (k Keeper) GetRewardRate(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	var rate sdk.Dec
	k.cdc.MustUnmarshalBinaryBare(store.Get([]byte(types.RewardRateKey)), &rate)
	return rate
}

// SetRewardRate store the vbr reward rate.
func (k Keeper) SetRewardRate(ctx sdk.Context, rate sdk.Dec) error {
	if err := types.ValidateRewardRate(rate); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.RewardRateKey), k.cdc.MustMarshalBinaryBare(rate))
	return nil
}

// IsDailyWighDrawBlock control if height is the daily withdraw block
func (k Keeper) IsDailyWighDrawBlock(height int64) bool {
	rest := height % (BPD.Int64() + 2)
	//rest := height % (10 + 2)
	if rest > 0 {
		return false
	}
	return true
}