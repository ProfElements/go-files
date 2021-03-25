package formats

/*
NAME:
EXTENSION:
DESCRIPTION:
*/

type File struct{}
type Work struct{}

func read(data []byte) (*File, error) { return nil, nil }
func write(*File) ([]byte, error)     { return nil, nil }
func decode(*File) *Work              { return nil }
func encode(*Work) *File              { return nil }
