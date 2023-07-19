package sl

import (
	"golang.org/x/exp/slog"
	_ "modernc.org/sqlite"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
