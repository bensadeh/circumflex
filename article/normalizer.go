package article

import "sort"

func normalizeHeaders(blocks []*block) {
	seen := make(map[int]bool)

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
		levels = append(levels, k)
	}

	sort.Ints(levels)

	mapping := make(map[int]int, len(levels))
	for i, level := range levels {
		mapping[level] = blockH1 + i
	}

	for _, b := range blocks {
		if isHeader(b.Kind) {
			b.Kind = mapping[b.Kind]
		}
	}
}

func isHeader(kind int) bool {
	return kind >= blockH1 && kind <= blockH6
}
