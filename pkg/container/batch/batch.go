package batch

func New(attrs []string) *Batch {
	return &Batch{
		Attrs: attrs,
		Vecs:  make([]any, len(attrs)),
	}
}

func SetLength(bat *Batch, n int) {
	/*
		for _, vec := range bat.Vecs {
			switch v := (interface{})(vec).(type) {
			case *vector.Vector[types.Int8]:
				vector.SetLength(v, n)
			case *vector.Vector[types.Int16]:
			case *vector.Vector[types.Int32]:
			case *vector.Vector[types.Int64]:
			case *vector.Vector[types.UInt8]:
			case *vector.Vector[types.UInt16]:
			case *vector.Vector[types.UInt32]:
			case *vector.Vector[types.UInt64]:
			}
		}
	*/
}
