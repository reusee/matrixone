package batch

func New(attrs []string) *Batch {
	return &Batch{
		Attrs: attrs,
		Vecs:  make([]any, len(attrs)),
	}
}
