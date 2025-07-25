package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgBlobMetadataTx{}

// ValidateBasic for this message should be a no-op so that we definitely AnteHandle this message.
// Events are being validated before composed as a transaction anyway.
func (*MsgBlobMetadataTx) ValidateBasic() error {
	return nil
}
