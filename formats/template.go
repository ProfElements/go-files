package formats

/*
NAME:
EXTENSION:
DESCRIPTION:
*/

import ()

type File struct{}
type Work struct{}

func read(data []byte) (*File, error) {}
func write(*File) ([]byte, error)     {}
func decode(*File) *Work              {}
func encode(*Work) *File              {}
