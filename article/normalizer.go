package article

import "sort"

func normalizeHeaders(blocks []*block) {
	seen := make(map[blockKind]bool)

	for _, b := range blocks {
		if isHeader(b.Kind) {
			seen[b.Kind] = true
		}
	}

	if len(seen) == 0 {
		return
	}

	levels := make([]int, 0, len(seen))
	for k := range seen {
		levels = append(levels, int(k))
	}

	sort.Ints(levels)

	mapping := make(map[blockKind]blockKind, len(levels))
	for i, level := range levels {
		mapping[blockKind(level)] = blockH1 + blockKind(i)
	}

	for _, b := range blocks {
		if isHeader(b.Kind) {
			b.Kind = mapping[b.Kind]
		}
	}
}

func isHeader(kind blockKind) bool {
	return kind >= blockH1 && kind <= blockH6
}
