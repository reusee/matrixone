package batch

// Batch represents a part of a relationship
//  (Attrs) - list of attributes
//  (vecs) 	- columns
type Batch struct {
	// Attrs column name list
	Attrs []string
	// Vecs col data
	Vecs []any
}
