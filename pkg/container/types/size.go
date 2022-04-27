package types

func (_ Bool) Size() int {
	return 1
}

func (_ Int8) Size() int {
	return 1
}

func (_ Int16) Size() int {
	return 2
}

func (_ Int32) Size() int {
	return 4
}

func (_ Int64) Size() int {
	return 8
}

func (_ UInt8) Size() int {
	return 1
}

func (_ UInt16) Size() int {
	return 2
}

func (_ UInt32) Size() int {
	return 4
}

func (_ UInt64) Size() int {
	return 8
}

func (_ Float32) Size() int {
	return 4
}

func (_ Float64) Size() int {
	return 8
}

func (_ Date) Size() int {
	return 4
}

func (_ Datetime) Size() int {
	return 8
}

func (_ Decimal64) Size() int {
	return 8
}

func (_ Decimal128) Size() int {
	return 16
}

func (b Bytes) Size() int {
	return len(b)
}
