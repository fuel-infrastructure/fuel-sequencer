package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
)

func NewAnysWithValue(v proto.Message) ([]*codectypes.Any, error) {
	anyMsg, err := codectypes.NewAnyWithValue(v)
	if err != nil {
		return nil, err
	}

	return []*codectypes.Any{anyMsg}, nil
}
