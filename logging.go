package hub

import (
	"bytes"
	"io"
)

var LogLevel string = "DEBUG"

var levels = []string{
	"DEBUG",
	"WARN",
	"ERROR",
}

type writerFunc func([]byte) (int, error)

func (w writerFunc) Write(b []byte) (int, error) {
	return w(b)
}

func leveledLogWriter(w io.Writer) io.Writer {
	lvls := make(map[string]int, len(levels))
	for i, l := range levels {
		lvls[l] = i
	}
	lowest := lvls[levels[0]]
	if l, found := lvls[LogLevel]; found {
		lowest = l
	}

	return writerFunc(func(b []byte) (int, error) {
		if start := bytes.IndexByte(b, byte('[')); start != -1 {
			if end := bytes.IndexByte(b, byte(']')); end != -1 {
				if level, ok := lvls[string(b[start+1:end])]; ok {
					if level >= lowest {
						return w.Write(b)
					}
					return len(b), nil
				}
			}
		}
		return w.Write(b)
	})
}
