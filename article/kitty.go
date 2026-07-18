package article

// KittyWork is one image whose terminal state is stale: PNG data still to
// transmit, or — with PNG nil — an already-transmitted image whose placement
// geometry changed with the layout.
type KittyWork struct {
	ID         int
	PNG        []byte
	Cols, Rows int
}

// PendingKittyWork returns the images the last render laid placeholder
// cells for whose terminal state no longer matches, and marks them clean —
// the caller owes the terminal exactly these sequences. Images the Kitty
// path never rendered (images hidden, terminal without graphics) want
// nothing and never appear.
func (p *Parsed) PendingKittyWork() []KittyWork {
	var work []KittyWork

	for i := range p.blocks {
		k := p.blocks[i].kitty
		if k == nil || k.wantCols == 0 {
			continue
		}

		switch {
		case !k.sent:
			work = append(work, KittyWork{ID: k.id, PNG: k.png, Cols: k.wantCols, Rows: k.wantRows})

		case k.wantCols != k.sentCols || k.wantRows != k.sentRows:
			work = append(work, KittyWork{ID: k.id, Cols: k.wantCols, Rows: k.wantRows})

		default:
			continue
		}

		k.sent = true
		k.sentCols, k.sentRows = k.wantCols, k.wantRows
	}

	return work
}
