package plugin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type EntryBaseTestSuite struct {
	suite.Suite
}

func (suite *EntryBaseTestSuite) TestNewEntry() {
	e := NewEntry("foo")
	assertOpTTL := func(op cacheableOp, opName string, expectedTTL time.Duration) {
		actualTTL := e.getTTLOf(op)
		suite.Equal(
			expectedTTL,
			actualTTL,
			"expected the TTL of %v to be %v, but got %v instead",
			opName,
			expectedTTL,
			actualTTL,
		)
	}

	suite.Equal("foo", e.Name())
	assertOpTTL(List, "List", 15*time.Second)
	assertOpTTL(Open, "Open", 15*time.Second)
	assertOpTTL(Metadata, "Metadata", 15*time.Second)

	e.setID("/foo")
	suite.Equal("/foo", e.ID())

	e.SetTTLOf(List, 40*time.Second)
	assertOpTTL(List, "List", 40*time.Second)

	e.TurnOffCachingFor(List)
	assertOpTTL(List, "List", -1)

	e.TurnOffCaching()
	assertOpTTL(Open, "Open", -1)
	assertOpTTL(Metadata, "Metadata", -1)

}

func TestEntryBase(t *testing.T) {
	suite.Run(t, new(EntryBaseTestSuite))
}