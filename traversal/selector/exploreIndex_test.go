package selector

import (
	"fmt"
	"testing"

	. "github.com/warpfork/go-wish"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

func TestParseExploreIndex(t *testing.T) {
	t.Run("parsing non map node should error", func(t *testing.T) {
		sn := basicnode.NewInt(0)
		_, err := ParseContext{}.ParseExploreIndex(sn)
		Wish(t, err, ShouldEqual, fmt.Errorf("selector spec parse rejected: selector body must be a map"))
	})
	t.Run("parsing map node without next field should error", func(t *testing.T) {
		sn := fluent.MustBuildMap(basicnode.Style__Map{}, 1, func(na fluent.MapAssembler) {
			na.AssembleEntry(SelectorKey_Index).AssignInt(2)
		})
		_, err := ParseContext{}.ParseExploreIndex(sn)
		Wish(t, err, ShouldEqual, fmt.Errorf("selector spec parse rejected: next field must be present in ExploreIndex selector"))
	})
	t.Run("parsing map node without index field should error", func(t *testing.T) {
		sn := fluent.MustBuildMap(basicnode.Style__Map{}, 1, func(na fluent.MapAssembler) {
			na.AssembleEntry(SelectorKey_Next).CreateMap(1, func(na fluent.MapAssembler) {
				na.AssembleEntry(SelectorKey_Matcher).CreateMap(0, func(na fluent.MapAssembler) {})
			})
		})
		_, err := ParseContext{}.ParseExploreIndex(sn)
		Wish(t, err, ShouldEqual, fmt.Errorf("selector spec parse rejected: index field must be present in ExploreIndex selector"))
	})
	t.Run("parsing map node with index field that is not an int should error", func(t *testing.T) {
		sn := fluent.MustBuildMap(basicnode.Style__Map{}, 2, func(na fluent.MapAssembler) {
			na.AssembleEntry(SelectorKey_Index).AssignString("cheese")
			na.AssembleEntry(SelectorKey_Next).CreateMap(1, func(na fluent.MapAssembler) {
				na.AssembleEntry(SelectorKey_Matcher).CreateMap(0, func(na fluent.MapAssembler) {})
			})
		})
		_, err := ParseContext{}.ParseExploreIndex(sn)
		Wish(t, err, ShouldEqual, fmt.Errorf("selector spec parse rejected: index field must be a number in ExploreIndex selector"))
	})
	t.Run("parsing map node with next field with invalid selector node should return child's error", func(t *testing.T) {
		sn := fluent.MustBuildMap(basicnode.Style__Map{}, 2, func(na fluent.MapAssembler) {
			na.AssembleEntry(SelectorKey_Index).AssignInt(2)
			na.AssembleEntry(SelectorKey_Next).AssignInt(0)
		})
		_, err := ParseContext{}.ParseExploreIndex(sn)
		Wish(t, err, ShouldEqual, fmt.Errorf("selector spec parse rejected: selector is a keyed union and thus must be a map"))
	})
	t.Run("parsing map node with next field with valid selector node should parse", func(t *testing.T) {
		sn := fluent.MustBuildMap(basicnode.Style__Map{}, 2, func(na fluent.MapAssembler) {
			na.AssembleEntry(SelectorKey_Index).AssignInt(2)
			na.AssembleEntry(SelectorKey_Next).CreateMap(1, func(na fluent.MapAssembler) {
				na.AssembleEntry(SelectorKey_Matcher).CreateMap(0, func(na fluent.MapAssembler) {})
			})
		})
		s, err := ParseContext{}.ParseExploreIndex(sn)
		Wish(t, err, ShouldEqual, nil)
		Wish(t, s, ShouldEqual, ExploreIndex{Matcher{}, [1]ipld.PathSegment{ipld.PathSegmentOfInt(2)}})
	})
}

func TestExploreIndexExplore(t *testing.T) {
	s := ExploreIndex{Matcher{}, [1]ipld.PathSegment{ipld.PathSegmentOfInt(3)}}
	t.Run("exploring should return nil unless node is a list", func(t *testing.T) {
		n := fluent.MustBuildMap(basicnode.Style__Map{}, 0, func(na fluent.MapAssembler) {})
		returnedSelector := s.Explore(n, ipld.PathSegmentOfInt(3))
		Wish(t, returnedSelector, ShouldEqual, nil)
	})
	n := fluent.MustBuildList(basicnode.Style__List{}, 4, func(na fluent.ListAssembler) {
		na.AssembleValue().AssignInt(0)
		na.AssembleValue().AssignInt(1)
		na.AssembleValue().AssignInt(2)
		na.AssembleValue().AssignInt(3)
	})
	t.Run("exploring should return nil when given a path segment with a different index", func(t *testing.T) {
		returnedSelector := s.Explore(n, ipld.PathSegmentOfInt(2))
		Wish(t, returnedSelector, ShouldEqual, nil)
	})
	t.Run("exploring should return nil when given a path segment that isn't an index", func(t *testing.T) {
		returnedSelector := s.Explore(n, ipld.PathSegmentOfString("cheese"))
		Wish(t, returnedSelector, ShouldEqual, nil)
	})
	t.Run("exploring should return the next selector when given a path segment with the right index", func(t *testing.T) {
		returnedSelector := s.Explore(n, ipld.PathSegmentOfInt(3))
		Wish(t, returnedSelector, ShouldEqual, Matcher{})
	})
}
