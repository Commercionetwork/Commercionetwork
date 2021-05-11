package types

import (
	"fmt"
	"testing"

	sdkErr "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var TestFunder, _ = sdk.AccAddressFromBech32("cosmos1lwmppctrr6ssnrmuyzu554dzf50apkfvd53jx0")
var TestAmount = sdk.NewCoins(sdk.Coin{
	Denom:  "ucommercio",
	Amount: sdk.NewInt(100),
})

var msgIncrementsBRPool = MsgIncrementBlockRewardsPool{
	Funder: TestFunder,
	Amount: TestAmount,
}

var msgIncrementsBrPoolNoFunds = MsgIncrementBlockRewardsPool{
	Funder: TestFunder,
	Amount: sdk.NewCoins(sdk.Coin{
		Denom:  "ucommercio",
		Amount: sdk.NewInt(0),
	}),
}

func TestMsgIncrementBlockRewardsPool_Route(t *testing.T) {
	actual := msgIncrementsBRPool.Route()
	expected := ModuleName

	require.Equal(t, expected, actual)
}

func TestMsgIncrementBlockRewardsPool_Type(t *testing.T) {
	actual := msgIncrementsBRPool.Type()
	expected := MsgTypeIncrementBlockRewardsPool

	require.Equal(t, expected, actual)
}

func TestMsgIncrementBlockRewardsPool_ValidateBasic_valid(t *testing.T) {
	actual := msgIncrementsBRPool.ValidateBasic()

	require.Nil(t, actual)
}

func TestMsgIncrementBlockRewardsPool_ValidateBasic_noFunds(t *testing.T) {
	actual := msgIncrementsBrPoolNoFunds.ValidateBasic()
	expected := sdkErr.Wrap(sdkErr.ErrUnknownRequest, "You can't transfer a null or negative amount")

	require.Equal(t, expected.Error(), actual.Error())
}

func TestMsgIncrementBlockRewardsPool_GetSignBytes(t *testing.T) {
	actual := msgIncrementsBRPool.GetSignBytes()
	expected := sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msgIncrementsBRPool))

	require.Equal(t, expected, actual)
}

func TestMsgIncrementBlockRewardsPool_GetSigners(t *testing.T) {
	actual := msgIncrementsBRPool.GetSigners()
	expected := []sdk.AccAddress{msgIncrementsBRPool.Funder}

	require.Equal(t, expected, actual)
}

// -------------------------
// --- MsgSetRewardRate
// -------------------------
var TestGovernement, _= sdk.AccAddressFromBech32("cosmos1lwmppctrr6ssnrmuyzu554dzf50apkfvd53jx0")
var TestRate = sdk.NewDec(100)
var TestNegativeRate = sdk.NewDec(-50)

var msgSetRewardRate = MsgSetRewardRate{
	Government: TestGovernement ,
	RewardRate: TestRate,
}

var msgSetRewardRateNegative = MsgSetRewardRate{
	Government: TestGovernement ,
	RewardRate: TestNegativeRate,
}

func TestMsgSetRewardRate_Route(t *testing.T){
	actual := msgSetRewardRate.Route()
	expected := ModuleName

	require.Equal(t, expected, actual)
}

func TestMsgSetRewardRate_Type(t *testing.T){
	actual := msgSetRewardRate.Type()
	expected := MsgTypeSetRewardRate

	require.Equal(t, expected, actual)
}

func TestMsgSetRewardRate_ValidateBasic(t *testing.T){
	actual := msgSetRewardRate.ValidateBasic()

	require.Nil(t, actual)
}

func TestMsgSetRewardRate_ValidateBasic_NegativeRate(t *testing.T){
	actual := msgSetRewardRateNegative.ValidateBasic()
	expected := fmt.Errorf("reward rate must be positive: %s", msgSetRewardRateNegative.RewardRate)

	require.Equal(t, expected.Error(), actual.Error())
}

func TestMsgSetRewardRate_GetSignBytes(t *testing.T){
	actual := msgSetRewardRate.GetSignBytes()
	expected := sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msgSetRewardRate))

	require.Equal(t, expected, actual)
}

func TestMsgSetRewardRate_GetSigners(t *testing.T){
	actual := msgSetRewardRate.GetSigners()
	expected := []sdk.AccAddress{msgSetRewardRate.Government}

	require.Equal(t, expected, actual)
}

// -------------------------
// --- MsgSetAutomaticWithdraw
// -------------------------
var TestAutomaticWithdraw bool

var msgSetAutomaticWithdraw = MsgSetAutomaticWithdraw{
	Government: TestGovernement,
	AutomaticWithdraw: TestAutomaticWithdraw,
}

func TestMsgSetAutomaticWithdraw_Route(t *testing.T){
	actual := msgSetAutomaticWithdraw.Route()
	expected := ModuleName

	require.Equal(t, expected, actual)
}

func TestMsgSetAutomaticWithdraw_type(t *testing.T){
	actual := msgSetAutomaticWithdraw.Type()
	expected := MsgTypeSetAutomaticWithdraw

	require.Equal(t, expected, actual)
}

func TestMsgSetAutomaticWithdraw_ValidateBasic(t *testing.T){
	actual := msgSetAutomaticWithdraw.ValidateBasic()
	
	require.Nil(t, actual)
}

func TestMsgSetAutomaticWithdraw_GetSignBytes(t *testing.T){
	actual := msgSetAutomaticWithdraw.GetSignBytes()
	expected := sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msgSetAutomaticWithdraw))

	require.Equal(t, expected, actual)
}

func TestMsgSetAutomaticWithdraw_GetSigners(t *testing.T){
	actual := msgSetAutomaticWithdraw.GetSigners()
	expected := []sdk.AccAddress{msgSetAutomaticWithdraw.Government}

	require.Equal(t, expected, actual)
}

func TestMsgSetAutomaticWithdraw(t *testing.T){
	//TODO..
}