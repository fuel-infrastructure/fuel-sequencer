package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
)

var _ sdk.Msg = &MsgBlobMetadataTx{}

// ValidateBasic for this message should be a no-op so that we definitely AnteHandle this message.
// Events are being validated before composed as a transaction anyway.
func (m *MsgBlobMetadataTx) ValidateBasic() error {
	if m == nil {
		return ErrMsgBlobMetadataTxNil
	}

	if m.Hash == "" {
		return ErrHashEmpty
	}

	if _, err := store.ParseKey(m.Hash); err != nil {
		return errors.Wrap(ErrHashNotValidKey, "hash: "+m.Hash)
	}

	if m.Size_ == 0 {
		return ErrSizeZero
	}

	if m.Topic == "" {
		return ErrTopicEmpty
	}

	return nil
}
