package meta

import (
	"testing"

	"github.com/puppetlabs/wash/cmd/internal/find/parser/parsertest"
	"github.com/puppetlabs/wash/cmd/internal/find/parser/predicate"
	"github.com/stretchr/testify/suite"
)

type PredicateTestSuite struct {
	parsertest.Suite
}

func (s *PredicateTestSuite) TestErrors() {
	s.RunTestCases(
		s.NPETC("", "expected either a primitive, object, or array predicate", true),
		// These cases ensure that parsePredicate returns any syntax errors found
		// while parsing the predicate
		s.NPETC(".", "expected a key sequence after '.'", false),
		s.NPETC("[", `expected a closing '\]'`, false),
		s.NPETC("--15", "positive", false),
		// These cases ensure that parsePredicate does not parse any expression operators.
		// Otherwise, parsePredicateExpression may not work correctly.
		s.NPETC("-a", ".*primitive.*", true),
		s.NPETC("-and", ".*primitive.*", true),
		s.NPETC("-o", ".*primitive.*", true),
		s.NPETC("-or", ".*primitive.*", true),
		s.NPETC("!", ".*primitive.*", true),
		s.NPETC("-not", ".*primitive.*", true),
		s.NPETC("(", ".*primitive.*", true),
		s.NPETC(")", ".*primitive.*", true),
	)
}

func (s *PredicateTestSuite) TestValidInput() {
	mp := make(map[string]interface{})
	mp["key"] = true
	s.RunTestCases(
		// ObjectPredicate
		s.NPTC(".key -true", "", mp),
		// ArrayPredicate
		s.NPTC("[?] -true", "", toA(true)),
		// PrimitivePredicate
		s.NPTC("-true", "", true),
	)
}

func TestPredicate(t *testing.T) {
	s := new(PredicateTestSuite)
	s.Parser = predicate.ToParser(parsePredicate)
	suite.Run(t, s)
}
