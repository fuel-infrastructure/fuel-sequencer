package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	utilstest "github.com/fuel-infrastructure/fuel-sequencer/testutil/utils"
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
		expTopic         *types.Topic
		expGasConsumed   uint64
		expErrMsg        string
	}{
		{
			name: "successfully post a blob - creates new topic",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic:      nil,
			setNonce:         math.ZeroInt(),
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			expGasConsumed: 12996, // Empty data gas: 12916 + 20 * 4 bytes = 12996 gas
			expErrMsg:      "",
		},
		{
			name: "successfully post a blob - nonce update verification",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(101),
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			setNonce:         math.NewInt(100),
			preSetTopic:      nil,
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			expGasConsumed: 13062, // Empty data gas: 12982 + 20 * 4 bytes = 13062 gas
			expErrMsg:      "",
		},
		{
			name: "successfully post a blob - updates existing topic",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.OneInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.OneInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			setNonce: math.ZeroInt(),
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: withdrawer,
				Order: math.OneInt(),
			},
			expGasConsumed: 13269, // Empty data gas: 13189 + 20 * 4 bytes = 13,269 gas (less gas topic already created)
			expErrMsg:      "",
		},
		{
			name: "successfully post a blob - large data",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 54),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 54),
			},
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic:      nil,
			setNonce:         math.ZeroInt(),
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			expGasConsumed: 13996, // Empty data gas: 12916 + 20 * 54 bytes = 13996 gas
			expErrMsg:      "",
		},
		{
			name: "post a blob that exceeds max size",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 500),
			},
			msgResponse:      nil,
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic:      nil,
			setNonce:         math.ZeroInt(),
			expTopic:         nil,
			expErrMsg:        "message size 500 exceeds max blob size bytes 400",
		},
		{
			name: "post a blob with incorrect order",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(1),
				Order: math.NewInt(2),
				Data:  []byte("data"),
			},
			msgResponse:      nil,
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(1),
				Owner: withdrawer,
				Order: math.ZeroInt(),
			},
			setNonce:  math.ZeroInt(),
			expTopic:  nil,
			expErrMsg: "msg order 2 doesn't match next topic order 1",
		},
		{
			name: "post a blob with mismatching topic owner",
			msg: types.MsgPostBlob{
				From:  withdrawer,
				Topic: utilstest.MockTopicIDHex(1),
				Order: math.OneInt(),
				Data:  []byte("data"),
			},
			msgResponse:      nil,
			maxBlobSizeBytes: 400,
			gasPerBlobByte:   20,
			preSetTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(1),
				Owner: anotherAccount,
				Order: math.ZeroInt(),
			},
			setNonce: math.ZeroInt(),
			expTopic: nil,
			expErrMsg: fmt.Sprintf(
				"from address %s doesn't match topic owner %s",
				withdrawer, anotherAccount,
			),
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
		})
	}
}
