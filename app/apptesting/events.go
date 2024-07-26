package apptesting

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/slices"
)

// AssertEventEmitted asserts that ctx's event manager has emitted the given number of events of the given type.
func (s *KeeperTestHelper) AssertEventEmitted(ctx sdk.Context, eventTypeExpected string, numEventsExpected int) {
	s.AssertEventInEventsList(ctx.EventManager().Events().ToABCIEvents(), eventTypeExpected, numEventsExpected)
}

// AssertEventInEventsList asserts that the events list argument has the given number of events of the given type.
func (s *KeeperTestHelper) AssertEventInEventsList(events []abci.Event, eventTypeExpected string, numEventsExpected int) {
	// filter out other events
	eventCounter := 0
	for _, event := range events {
		if event.Type == eventTypeExpected {
			eventCounter += 1
		}
	}
	s.Equal(numEventsExpected, eventCounter)
}

func (s *KeeperTestHelper) FindEvent(events []sdk.Event, name string) sdk.Event {
	index := slices.IndexFunc(events, func(e sdk.Event) bool { return e.Type == name })
	return events[index]
}

func (s *KeeperTestHelper) ExtractAttributes(event sdk.Event) map[string]string {
	attrs := make(map[string]string)
	for _, a := range event.Attributes {
		attrs[string(a.Key)] = string(a.Value)
	}
	return attrs
}
