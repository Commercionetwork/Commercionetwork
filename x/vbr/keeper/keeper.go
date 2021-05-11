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

	government "github.com/commercionetwork/commercionetwork/x/government/keeper"

	ctypes "github.com/commercionetwork/commercionetwork/x/common/types"
	"github.com/commercionetwork/commercionetwork/x/vbr/types"
)

// --------------------
// --- Blocks per day
// --- Blocks per year
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

// Keeper is keeper type
type Keeper struct {
	cdc          *codec.Codec
	storeKey     sdk.StoreKey
	distKeeper   distribution.Keeper
	supplyKeeper supply.Keeper
	govKeeper    government.Keeper
}

// NewKeeper create Keeper
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, dk distribution.Keeper, sk supply.Keeper, gk government.Keeper) Keeper {
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		distKeeper:   dk,
		supplyKeeper: sk,
		govKeeper:    gk,
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

// ---------------------------
// --- Reward distribution
// ---------------------------

// ComputeProposerReward computes the final reward for the validator block's proposer
func (k Keeper) ComputeProposerReward(ctx sdk.Context, vCount int64, proposer exported.ValidatorI, denom string) sdk.DecCoins {

	// Get rewarded rate
	rewardRate := k.GetRewardRate(ctx)

	// Calculate rewarded rate with validator percentage
	rewardRateVal := rewardRate.Mul(sdk.NewDec(vCount)).Quo(sdk.NewDec(100))

	// Get total bonded token of validator
	proposerBonded := proposer.GetBondedTokens()

	// Compute reward for each block
	return sdk.NewDecCoins(sdk.NewDecCoinFromDec(denom, proposerBonded.ToDec().Mul(rewardRateVal).Quo(BPD)))
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
		// TODO this error continue when pool hasn't enough funds for all rewards. Find a method to avoid this
		return sdkErr.Wrap(sdkErr.ErrInsufficientFunds, "Pool hasn't got enough funds to supply validator's rewards")
	}

	return nil
}

// WithdrawAllRewards withdraw reward to all validator
func (k Keeper) WithdrawAllRewards(ctx sdk.Context, stakeKeeper staking.Keeper) error {
	// Loop throw delegations and withdraw all rewards from each validator
	// and immediately delegate to validator
	dels := stakeKeeper.GetAllDelegations(ctx)
	for _, delegation := range dels {
		returnedCoins, err := k.distKeeper.WithdrawDelegationRewards(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)
		if err == nil {
			amountRedelegate := returnedCoins.AmountOf(stakeKeeper.BondDenom(ctx))
			if amountRedelegate.IsPositive() {
				curValidator, found := stakeKeeper.GetValidator(ctx, delegation.ValidatorAddress)
				if !found {
					continue
				}
				_, _ = stakeKeeper.Delegate(ctx, delegation.DelegatorAddress, amountRedelegate, sdk.Unbonded, curValidator, true)
			}
		}
	}

	// Loop throw validators and withdraw all commission
	// and immediately delegate to validator
	vals := stakeKeeper.GetAllValidators(ctx)
	for _, validator := range vals {
		returnedCommission, err := k.distKeeper.WithdrawValidatorCommission(ctx, validator.GetOperator())

		if err == nil {
			amountRedelegate := returnedCommission.AmountOf(stakeKeeper.BondDenom(ctx))
			if amountRedelegate.IsPositive() {
				_, _ = stakeKeeper.Delegate(ctx, sdk.AccAddress(validator.GetOperator()), amountRedelegate, sdk.Unbonded, validator, true)
			}
		}
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

// GetAutomaticWithdraw retrieve automatic withdraw flag.
func (k Keeper) GetAutomaticWithdraw(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	var autoW bool
	k.cdc.MustUnmarshalBinaryBare(store.Get([]byte(types.AutomaticWithdraw)), &autoW)
	return autoW
}

// SetAutomaticWithdraw store the automatic withdraw flag.
func (k Keeper) SetAutomaticWithdraw(ctx sdk.Context, autoW bool) error {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.AutomaticWithdraw), k.cdc.MustMarshalBinaryBare(autoW))
	return nil
}

// IsDailyWithdrawBlock control if height is the daily withdraw block
// BPD.Int64() return -8061083817814786048
// But it would have to return 12960 (--> HoursPerDay * MinutesPerHour * BlocksPerMinute --> 24 * 60 * 9 )
// Using RoundInt64() instead of Int64() it's work but RoundInt64() panics if something goes wrong...we may not use it
func (k Keeper) IsDailyWithdrawBlock(height int64) bool {
	if height == 0 {
		return false
	}

	rest := height % BPD.Int64()
	
	return rest == 0
}
