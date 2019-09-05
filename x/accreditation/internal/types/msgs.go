package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// --------------------------
// --- MsgSetAccrediter
// --------------------------

// MsgSetAccrediter should be used when wanting to set a specific accrediter
// for a specific user.
// Note that the associated signer should be a trustworthy one in order to avoid
// unauthorized users to perform such assignment.
type MsgSetAccrediter struct {
	User       sdk.AccAddress `json:"user"`
	Accrediter sdk.AccAddress `json:"accrediter"`
	Signer     sdk.AccAddress `json:"signer"`
}

func NewMsgSetAccrediter(user, accrediter, signer sdk.AccAddress) MsgSetAccrediter {
	return MsgSetAccrediter{
		User:       user,
		Accrediter: accrediter,
		Signer:     signer,
	}
}

// RouterKey Implements Msg.
func (msg MsgSetAccrediter) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSetAccrediter) Type() string { return MsgTypeSetAccrediter }

// ValidateBasic Implements Msg.
func (msg MsgSetAccrediter) ValidateBasic() sdk.Error {
	if msg.User.Empty() {
		return sdk.ErrInvalidAddress(msg.User.String())
	}
	if msg.Accrediter.Empty() {
		return sdk.ErrInvalidAddress(msg.Accrediter.String())
	}
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSetAccrediter) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgSetAccrediter) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// --------------------------
// --- MsgDistributeReward
// --------------------------

// MsgDistributeReward should be used when wanting to distribute a
// specific reward to an accrediter verifying that he has previously
// accreditated all the users.
// Note that the signer should be a trustworthy one in order to avoid
// unauthorized reward distributions.
type MsgDistributeReward struct {
	Accrediter sdk.AccAddress   `json:"accrediter"`
	Users      []sdk.AccAddress `json:"users"`
	Signer     sdk.AccAddress   `json:"signer"`
	Reward     sdk.Coins        `json:"reward"`
}

// RouterKey Implements Msg.
func (msg MsgDistributeReward) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgDistributeReward) Type() string { return MsgTypeDistributeReward }

// ValidateBasic Implements Msg.
func (msg MsgDistributeReward) ValidateBasic() sdk.Error {
	if msg.Accrediter.Empty() {
		return sdk.ErrInvalidAddress(msg.Accrediter.String())
	}
	if len(msg.Users) == 0 {
		return sdk.ErrUnknownRequest("users cannot be empty")
	}
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.Reward.Empty() {
		return sdk.ErrUnknownRequest("reward cannot be empty")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgDistributeReward) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgDistributeReward) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
