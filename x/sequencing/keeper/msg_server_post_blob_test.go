package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func (s *KeeperTestSuite) TestPostBlob() {
	withdrawer := s.TestAccs[0].String()
	anotherAccount := s.TestAccs[1].String()
	testCases := []struct {
		name             string
		msg              types.MsgPostBlob
		msgResponse      *types.MsgPostBlobResponse
		maxBlobSizeBytes uint64
		gasPerBlobByte   uint64
		preSetTopic      *types.Topic
		setNonce         math.Int
		expTopicId       math.Int
		expTopic         *types.Topic
		expGasConsumed   uint64
		expErrMsg        string
	}{
		{
			name: "successfully post a blob - creates new topic",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic:      nil,
			setNonce:         math.ZeroInt(),
			expTopicId:       math.OneInt(),
			expTopic: &types.Topic{
				Id:    math.ZeroInt(),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			expGasConsumed: 13644, // Empty data gas: 14621 + 20 * 4 bytes = 13644 gas
			expErrMsg:      "",
		},
		{
			name: "successfully post a blob - nonce update verification",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(101),
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			setNonce:         math.NewInt(100),
			preSetTopic:      nil,
			expTopicId:       math.OneInt(),
			expTopic: &types.Topic{
				Id:    math.ZeroInt(),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			expGasConsumed: 13710, // Empty data gas: 13630 + 20 * 4 bytes = 13710 gas
			expErrMsg:      "",
		},
		{
			name: "successfully post a blob - updates existing topic",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.OneInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.OneInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic: &types.Topic{
				Id:    math.ZeroInt(),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			setNonce:   math.ZeroInt(),
			expTopicId: math.OneInt(),
			expTopic: &types.Topic{
				Id:    math.ZeroInt(),
				Owner: withdrawer,
				Order: math.OneInt(),
			},
			expGasConsumed: 10200, // Empty data gas: 10120 + 20 * 4 bytes = 10200 gas (less gas topic already created)
			expErrMsg:      "",
		},
		{
			name: "successfully post a blob - large data",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.ZeroInt(),
				Data:  make([]byte, 54),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.ZeroInt(),
				Data:  make([]byte, 54),
			},
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic:      nil,
			setNonce:         math.ZeroInt(),
			expTopicId:       math.OneInt(),
			expTopic: &types.Topic{
				Id:    math.ZeroInt(),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			expGasConsumed: 14644, // Empty data gas: 13564 + 20 * 54 bytes = 14644 gas
			expErrMsg:      "",
		},
		{
			name: "post a blob that exceeds max size",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: math.ZeroInt(),
				Order: math.ZeroInt(),
				Data:  make([]byte, 500),
			},
			msgResponse:      nil,
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic:      nil,
			setNonce:         math.ZeroInt(),
			expTopicId:       math.ZeroInt(),
			expTopic:         nil,
			expErrMsg:        "message size 500 exceeds max blob size bytes 400",
		},
		{
			name: "post a blob with incorrect order",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: math.OneInt(),
				Order: math.NewInt(2),
				Data:  []byte("data"),
			},
			msgResponse:      nil,
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic: &types.Topic{
				Id:    math.OneInt(),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			setNonce:   math.ZeroInt(),
			expTopicId: math.NewInt(2),
			expTopic:   nil,
			expErrMsg:  "msg order 2 doesn't match next topic order 1",
		},
		{
			name: "post a blob with mismatching topic owner",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: math.OneInt(),
				Order: math.OneInt(),
				Data:  []byte("data"),
			},
			msgResponse:      nil,
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic: &types.Topic{
				Id:    math.OneInt(),
				Owner: anotherAccount,
				Order: math.ZeroInt(),
			},
			setNonce:   math.ZeroInt(),
			expTopicId: math.NewInt(2),
			expTopic:   nil,
			expErrMsg: fmt.Sprintf(
				"from address %s doesn't match topic owner %s",
				withdrawer, anotherAccount,
			),
		},
		{
			name: "post a blob with mismatching topic id",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: math.NewInt(2),
				Order: math.OneInt(),
				Data:  []byte("data"),
			},
			msgResponse:      nil,
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic: &types.Topic{
				Id:    math.ZeroInt(),
				Owner: anotherAccount,
				Order: math.ZeroInt(),
			},
			setNonce:   math.ZeroInt(),
			expTopicId: math.OneInt(),
			expTopic:   nil,
			expErrMsg:  "msg topic 2 doesn't match next topic id 1",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.SequencingKeeper)

			// Set sequencing parameters
			params := types.DefaultParams()
			params.MaxBlobSizeBytes = tc.maxBlobSizeBytes
			params.GasPerBlobByte = tc.gasPerBlobByte

			_ = s.App.SequencingKeeper.SetParams(s.Ctx(), params)

			// Pre-set a topic if needed as well as the next topic id
			if tc.preSetTopic != nil {
				s.App.SequencingKeeper.SetTopic(s.Ctx(), *tc.preSetTopic)
				s.App.SequencingKeeper.SetNextTopicId(s.Ctx(), tc.preSetTopic.Id.Add(math.OneInt()))
			} else {
				s.App.SequencingKeeper.SetNextTopicId(s.Ctx(), math.ZeroInt())
			}

			// Set Eth nonce
			if tc.setNonce.GT(math.ZeroInt()) {
				s.App.BridgeKeeper.SetLastEthereumNonce(s.Ctx(), tc.setNonce)
			}

			gasCtx := s.Ctx()
			response, err := msgServer.PostBlob(gasCtx, &tc.msg)
			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)

				// Verify Nonce was set
				lastNonce := s.App.BridgeKeeper.MustGetLastEthereumNonce(s.Ctx())
				s.Require().True(lastNonce.Equal(tc.msgResponse.Nonce))

			}

			s.Require().Equal(tc.msgResponse, response)

			// Check if the gas consumed matches
			if tc.expGasConsumed > 0 {
				gasConsumed := gasCtx.GasMeter().GasConsumed()
				s.Require().Equal(tc.expGasConsumed, gasConsumed)
			}

			// Verify the topic has been updated
			if tc.expTopic != nil {
				topic, found := s.App.SequencingKeeper.GetTopic(s.Ctx(), tc.expTopic.Id)
				s.Require().True(found)
				s.Require().Equal(tc.expTopic, &topic)
			}

			// Verify the next topic id is as expected
			actTopicId := s.App.SequencingKeeper.MustGetNextTopicId(s.Ctx())
			s.Require().Equal(tc.expTopicId, actTopicId)
		})
	}
}
