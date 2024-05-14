package keeper_test

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	utilstest "github.com/fuel-infrastructure/fuel-sequencer/testutil/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func (s *KeeperTestSuite) TestPostBlob() {

	sender := s.TestAccs[0].String()
	senderUpper := strings.ToUpper(s.TestAccs[0].String())
	senderLower := strings.ToLower(s.TestAccs[0].String())
	anotherAccount := s.TestAccs[1].String()

	testCases := []struct {
		name             string
		msg              types.MsgPostBlob
		msgResponse      *types.MsgPostBlobResponse
		maxBlobSizeBytes uint64
		preSetTopic      *types.Topic
		setNonce         math.Int
		expTopic         *types.Topic
		expErrMsg        string
	}{
		{
			name: "successfully post a blob - creates new topic - lowercase sender",
			msg: types.MsgPostBlob{
				From:  senderLower,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  senderLower,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			setNonce:         math.ZeroInt(),
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: senderLower,
				Order: math.ZeroInt(),
			},
			expErrMsg: "",
		},
		{
			name: "successfully post a blob - creates new topic - uppercase sender",
			msg: types.MsgPostBlob{
				From:  senderUpper,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  senderLower, // changed to lowercase
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			setNonce:         math.ZeroInt(),
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: senderUpper,
				Order: math.ZeroInt(),
			},
			expErrMsg: "",
		},
		{
			name: "successfully post a blob - nonce update verification",
			msg: types.MsgPostBlob{
				From:  sender,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(101),
				From:  sender,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			setNonce:         math.NewInt(100),
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: sender,
				Order: math.ZeroInt(),
			},
			expErrMsg: "",
		},
		{
			name: "successfully post a blob - updates existing topic",
			msg: types.MsgPostBlob{
				From:  sender,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.OneInt(),
				Data:  make([]byte, 4),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  sender,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.OneInt(),
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			preSetTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: sender,
				Order: math.ZeroInt(),
			},
			setNonce: math.ZeroInt(),
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: sender,
				Order: math.OneInt(),
			},
			expErrMsg: "",
		},
		{
			name: "successfully post a blob - large data",
			msg: types.MsgPostBlob{
				From:  sender,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 54),
			},
			msgResponse: &types.MsgPostBlobResponse{
				Nonce: math.NewInt(1),
				From:  sender,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 54),
			},
			maxBlobSizeBytes: 400,
			setNonce:         math.ZeroInt(),
			expTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(0),
				Owner: sender,
				Order: math.ZeroInt(),
			},
			expErrMsg: "",
		},
		{
			name: "post a blob that exceeds max size",
			msg: types.MsgPostBlob{
				From:  sender,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.ZeroInt(),
				Data:  make([]byte, 500),
			},
			maxBlobSizeBytes: 400,
			setNonce:         math.ZeroInt(),
			expErrMsg:        "message size 500 exceeds max blob size bytes 400",
		},
		{
			name: "post a blob with incorrect order",
			msg: types.MsgPostBlob{
				From:  sender,
				Topic: utilstest.MockTopicIDHex(1),
				Order: math.NewInt(2),
				Data:  []byte("data"),
			},
			maxBlobSizeBytes: 400,
			preSetTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(1),
				Owner: sender,
				Order: math.ZeroInt(),
			},
			setNonce:  math.ZeroInt(),
			expErrMsg: "msg order 2 doesn't match next topic order 1",
		},
		{
			name: "post a new blob with non-zero order",
			msg: types.MsgPostBlob{
				From:  sender,
				Topic: utilstest.MockTopicIDHex(0),
				Order: math.OneInt(), // non-zero
				Data:  make([]byte, 4),
			},
			maxBlobSizeBytes: 400,
			setNonce:         math.ZeroInt(),
			expErrMsg:        "msg order 1 expected to be 0 for new topics",
		},
		{
			name: "post a blob with mismatching topic owner",
			msg: types.MsgPostBlob{
				From:  sender,
				Topic: utilstest.MockTopicIDHex(1),
				Order: math.OneInt(),
				Data:  []byte("data"),
			},
			maxBlobSizeBytes: 400,
			preSetTopic: &types.Topic{
				Id:    utilstest.MockTopicIDHex(1),
				Owner: anotherAccount,
				Order: math.ZeroInt(),
			},
			setNonce: math.ZeroInt(),
			expErrMsg: fmt.Sprintf(
				"from address %s doesn't match topic owner %s",
				sender, anotherAccount,
			),
		},
		{
			name: "post a blob with invalid topic owner",
			msg: types.MsgPostBlob{
				From:  "some-invalid-address", // invalid!
				Topic: utilstest.MockTopicIDHex(1),
				Order: math.ZeroInt(),
				Data:  []byte("data"),
			},
			maxBlobSizeBytes: 400,
			setNonce:         math.ZeroInt(),
			expErrMsg:        "invalid topic address: decoding bech32 failed",
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
				s.Require().Nil(response)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.msgResponse, response)

			// Verify Nonce was set
			lastNonce := s.App.BridgeKeeper.MustGetLastEthereumNonce(s.Ctx())
			s.Require().True(lastNonce.Equal(tc.msgResponse.Nonce))

			// Verify the topic has been updated
			topic, found := s.App.SequencingKeeper.GetTopic(s.Ctx(), tc.expTopic.Id)
			s.Require().True(found)
			s.Require().Equal(tc.expTopic, &topic)
		})
	}
}
