// This file is a modified version of Go's net/http/pattern.go, copyright 2023 The Go Authors,
// which is licensed under the BSD 3-Clause License.
//
// The syntax and behavior of path patterns is described in docs/rfcs/2024-06-03-rate-limiting.md
//
// Notable changes from the stdlib:
// - GET and HEAD are not treated as the same for matching purposes
// - no support for matching the host
// - no implicit prefix matching
// - no explicit "end of string" matching ("{$}")
// - simpler path matching (we're not building a graph of routing nodes)

package ratelimiter

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
)

// A pattern is something that can be matched against an HTTP request.
// It has an optional method and a path.
type pattern struct {
	// The original pattern string
	pattern string
	// Optional method
	method string
	// Segments representing the parsed pattern
	segments []segment
}

// A segment matches one or more path segments.
//
// If `identifier` is false, it matches a literal segment.
// Example:
// - "foo" => segment{s:"foo"}
//
// If `identifier` is true and `multi` is false, it matches a single path segment.
// Example:
// - "{foo}" => segment{s:"foo", identifier:true}
//
// If `identifier` is true and `multi` is true, it matches all remaining path segments.
// Example:
// - "{all...}" => segment{s:"all", identifier:true, multi:true}
type segment struct {
	// Either the literal or the name of an identifier
	s string
	// Whether this segment is an identifier
	identifier bool
	// Whether this segment is an identifier that matches across segments
	multi bool
}

// Zero-sized value used in sets.
type empty struct{}

// newPattern parses a pattern string and returns a pointer to a pattern.
func newPattern(p string) (*pattern, error) {
	if len(p) == 0 {
		return nil, errors.New("empty pattern")
	}

	method, path, found := strings.Cut(p, " ")
	if !found {
		path = method
		method = ""
	} else if err := validateMethod(method); err != nil {
		return nil, err
	}

	pattern := &pattern{
		pattern: p,
		method:  method,
	}

	i := strings.IndexByte(path, '/')
	if i < 0 {
		return nil, errors.New("path missing a slash ('/')")
	}

	seenNames := map[string]empty{}
	for len(path) > 0 {
		// Invariant: path[0] == '/'
		path = path[1:]
		if len(path) == 0 {
			// Trailing slash
			break
		}

		i := strings.IndexByte(path, '/')
		if i < 0 {
			i = len(path)
		}
		var seg string
		seg, path = path[:i], path[i:]

		i = strings.IndexByte(seg, '{')
		if i < 0 {
			// Literal
			pattern.segments = append(pattern.segments, segment{s: seg})
			continue
		}

		// An identifier - either single or multi
		if i != 0 {
			return nil, errors.New("bad identifier segment (must start with '{')")
		}
		if seg[len(seg)-1] != '}' {
			return nil, errors.New("bad identifier segment (must end with '}')")
		}
		seg = seg[1 : len(seg)-1]

		name, multi := strings.CutSuffix(seg, "...")
		if multi && len(path) != 0 {
			return nil, errors.New("{...} identifier not at the end")
		}
		if name == "" {
			return nil, errors.New("empty identifier")
		}
		if _, ok := seenNames[name]; ok {
			return nil, clouderr.WithFields("duplicate identifier name", zap.String("name", name))
		}

		seenNames[name] = empty{}
		pattern.segments = append(pattern.segments, segment{s: name, identifier: true, multi: multi})
	}

	return pattern, nil
}

func validateMethod(method string) error {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions:
		return nil
	default:
		return clouderr.WithFields("invalid method", zap.String("method", method))
	}
}

func (p *pattern) lastSegment() segment {
	return p.segments[len(p.segments)-1]
}

// relationship is a relationship between two patterns, p1 and p2.
type relationship string

const (
	equivalent   relationship = "equivalent"   // both match the same requests
	moreGeneral  relationship = "moreGeneral"  // p1 matches everything p2 does & more
	moreSpecific relationship = "moreSpecific" // p2 matches everything p1 does & more
	disjoint     relationship = "disjoint"     // there is no request both patterns match
	overlaps     relationship = "overlaps"     // there is a request both patterns match, but neither is more specific
)

func (p *pattern) comparePathsAndMethods(p2 *pattern) relationship {
	mrel := p.compareMethods(p2)
	if mrel == disjoint {
		return disjoint
	}

	prel := p.comparePaths(p2)
	return combineRelationships(mrel, prel)
}

// compareMethods determines the relationship between the method
// part of patterns p and p2.
//
// A method can either be empty, "GET", or something else.
// The empty string matches any method, so it is the most general.
// Anything else matches only itself.
func (p *pattern) compareMethods(p2 *pattern) relationship {
	if p.method == p2.method {
		return equivalent
	}

	if p.method == "" {
		return moreGeneral
	}

	if p2.method == "" {
		return moreSpecific
	}

	return disjoint
}

// comparePaths determines the relationship between the path
// part of two patterns.
func (p *pattern) comparePaths(p2 *pattern) relationship {
	// If a path pattern doesn't end in a multi ("...") identifier, then it
	// can only match paths with the same number of segments.
	if len(p.segments) != len(p2.segments) && !p.lastSegment().multi && !p2.lastSegment().multi {
		return disjoint
	}

	// Consider corresponding segments in the two path patterns.
	var segs1, segs2 []segment
	rel := equivalent
	for segs1, segs2 = p.segments, p2.segments; len(segs1) > 0 && len(segs2) > 0; segs1, segs2 = segs1[1:], segs2[1:] {
		rel = combineRelationships(rel, compareSegments(segs1[0], segs2[0]))
		if rel == disjoint {
			return rel
		}
	}
	// We've reached the end of the corresponding patterns' segments.
	// If they have the same number of segments, then we've already determined
	// their relationship.
	if len(segs1) == 0 && len(segs2) == 0 {
		return rel
	}

	// Otherwise, the only way they could fail to be disjoint is if the shorter
	// pattern ends in a multi. In that case, that multi is more general
	// than the remainder of the longer pattern, so combine those two relationships.
	if len(segs1) < len(segs2) && p.lastSegment().multi {
		return combineRelationships(rel, moreGeneral)
	}
	if len(segs2) < len(segs1) && p2.lastSegment().multi {
		return combineRelationships(rel, moreSpecific)
	}

	return disjoint
}

// compareSegments determines the relationship between two segments.
func compareSegments(s1, s2 segment) relationship {
	if s1.multi && s2.multi {
		return equivalent
	}
	if s1.multi {
		return moreGeneral
	}
	if s2.multi {
		return moreSpecific
	}

	if s1.identifier && s2.identifier {
		return equivalent
	}
	if s1.identifier {
		return moreGeneral
	}
	if s2.identifier {
		return moreSpecific
	}

	// Both literals.
	if s1.s == s2.s {
		return equivalent
	}

	return disjoint
}

// combineRelationships determines the overall relationship of two patterns
// given the relationships of a partition of the patterns into two parts.
//
// For example, if p1 is more general than p2 in one way but equivalent
// in the other, then it is more general overall.
//
// Or if p1 is more general in one way and more specific in the other, then
// they overlap.
func combineRelationships(r1, r2 relationship) relationship {
	switch r1 {
	case equivalent:
		return r2
	case disjoint:
		return disjoint
	case overlaps:
		if r2 == disjoint {
			return disjoint
		}
		return overlaps
	case moreGeneral, moreSpecific:
		switch r2 {
		case equivalent:
			return r1
		case inverseRelationship(r1):
			return overlaps
		default:
			return r2
		}
	default:
		panic(fmt.Sprintf("unknown relationship %q", r1))
	}
}

// If p1 has relationship `r` to p2, then
// p2 has inverseRelationship(r) to p1.
func inverseRelationship(r relationship) relationship {
	switch r {
	case moreSpecific:
		return moreGeneral
	case moreGeneral:
		return moreSpecific
	default:
		return r
	}
}

// match reports whether the pattern matches the method and path segments.
func (p *pattern) match(method string, segments []string) bool {
	// If the pattern has a method, it must match the request's method.
	if p.method != "" && p.method != method {
		return false
	}

	if len(segments) != len(p.segments) {
		// If the last segment in the path is not a multi and there's a different number of segments, it can't match
		if !p.lastSegment().multi {
			return false
		}

		// Even if it's multi, if there are fewer segments than the pattern, it can't match
		if len(segments) < len(p.segments) {
			return false
		}
	}

	// We now know that the lengths are either equal, or that `segments` is longer,
	// but the last segment in the pattern is a multi.
	for i, seg := range p.segments {
		// Multi must be at the end of the pattern - if we're here, we match everything,
		// regardless of how many segments are left.
		if seg.multi {
			return true
		}

		// If the segment is an identifier, it will match anything in that segment, but not across them - keep looking.
		if seg.identifier {
			continue
		}

		// Otherwise, we know this segment is a literal - it must match the corresponding segment in the path.
		if segments[i] != seg.s {
			return false
		}
	}

	return true
}
