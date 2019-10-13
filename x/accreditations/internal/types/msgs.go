package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// --------------------------
// --- MsgInviteUser
// --------------------------

// MsgInviteUser allows to properly invite a user.
// Te invitation system should be a one-invite-only system, where invites
// consecutive to the first one should be discarded.
type MsgInviteUser struct {
	Recipient sdk.AccAddress `json:"recipient"`
	Sender    sdk.AccAddress `json:"sender"`
}

func NewMsgInviteUser(recipient, sender sdk.AccAddress) MsgInviteUser {
	return MsgInviteUser{
		Recipient: recipient,
		Sender:    sender,
	}
}

// Route Implements Msg.
func (msg MsgInviteUser) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgInviteUser) Type() string { return MsgTypeInviteUser }

// ValidateBasic Implements Msg.
func (msg MsgInviteUser) ValidateBasic() sdk.Error {
	if msg.Recipient.Empty() {
		return sdk.ErrInvalidAddress(msg.Recipient.String())
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgInviteUser) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgInviteUser) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// --------------------------
// --- MsgSetUserVerified
// --------------------------

// MsgSetUserVerified is used to set a specific user as properly verified.
// Note that the verifier address should identify a Trusted Service Provider account.
type MsgSetUserVerified struct {
	Timestamp time.Time      `json:"timestamp"` // Timestamp of the verification
	User      sdk.AccAddress `json:"user"`      // Recipient that has been verified
	Verifier  sdk.AccAddress `json:"verifier"`  // Trusted Service Provider
}

func NewMsgSetUserVerified(user sdk.AccAddress, timestamp time.Time, verifier sdk.AccAddress) MsgSetUserVerified {
	return MsgSetUserVerified{
		Timestamp: timestamp,
		User:      user,
		Verifier:  verifier,
	}
}

// Route Implements Msg.
func (msg MsgSetUserVerified) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSetUserVerified) Type() string { return MsgTypeSetUserVerified }

// ValidateBasic Implements Msg.
func (msg MsgSetUserVerified) ValidateBasic() sdk.Error {
	if msg.Timestamp.IsZero() {
		return sdk.ErrUnknownRequest("Timestamp not valid")
	}
	if msg.User.Empty() {
		return sdk.ErrInvalidAddress(msg.User.String())
	}
	if msg.Verifier.Empty() {
		return sdk.ErrInvalidAddress(msg.Verifier.String())
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSetUserVerified) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgSetUserVerified) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Verifier}
}

// --------------------------------
// --- MsgDepositIntoLiquidityPool
// --------------------------------

// MsgDepositIntoLiquidityPool should be used when wanting to deposit a specific
// amount into the liquidity pool which contains the total amount of rewards to
// be distributed upon an accreditation
type MsgDepositIntoLiquidityPool struct {
	Depositor sdk.AccAddress `json:"depositor"`
	Amount    sdk.Coins      `json:"amount"`
}

func NewMsgDepositIntoLiquidityPool(amount sdk.Coins, depositor sdk.AccAddress) MsgDepositIntoLiquidityPool {
	return MsgDepositIntoLiquidityPool{
		Depositor: depositor,
		Amount:    amount,
	}
}

// Route Implements Msg.
func (msg MsgDepositIntoLiquidityPool) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgDepositIntoLiquidityPool) Type() string { return MsgTypesDepositIntoLiquidityPool }

// ValidateBasic Implements Msg.
func (msg MsgDepositIntoLiquidityPool) ValidateBasic() sdk.Error {
	if msg.Depositor.Empty() {
		return sdk.ErrInvalidAddress(msg.Depositor.String())
	}
	if msg.Amount.Empty() {
		return sdk.ErrUnknownRequest("amount cannot be empty")
	}
	if msg.Amount.IsAnyNegative() {
		return sdk.ErrUnknownRequest("amount cannot be negative")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgDepositIntoLiquidityPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgDepositIntoLiquidityPool) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// --------------------------------
// --- MsgAddTsp
// --------------------------------

// MsgAddTsp should be used when wanting to add a new address
// as a Trusted Service Provider (TSP).
// TSPs will be able to sign rewarding-giving transactions
// so should be only a handful of very trusted accounts.
type MsgAddTsp struct {
	Tsp        sdk.AccAddress `json:"tsp"`
	Government sdk.AccAddress `json:"government"`
}

// Route Implements Msg.
func (msg MsgAddTsp) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgAddTsp) Type() string { return MsgTypeAddTsp }

// ValidateBasic Implements Msg.
func (msg MsgAddTsp) ValidateBasic() sdk.Error {
	if msg.Tsp.Empty() {
		return sdk.ErrInvalidAddress(msg.Tsp.String())
	}
	if msg.Government.Empty() {
		return sdk.ErrInvalidAddress(msg.Government.String())
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgAddTsp) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgAddTsp) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Government}
}
