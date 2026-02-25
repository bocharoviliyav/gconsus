package logging

import (
	"log/slog"
	"os"
	"strings"
)

func InitLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	replace := func(groups []string, a slog.Attr) slog.Attr {
		// delete date and time
		if a.Key == slog.TimeKey && len(groups) == 0 {
			return slog.Attr{}
		}

		// delete path prefix
		if a.Key == slog.SourceKey {
			s := a.Value.Any().(*slog.Source)
			splitted := strings.SplitN(s.File, string(os.PathSeparator), 2)

			if len(splitted) == 1 {
				s.File = splitted[0]
			} else {
				s.File = splitted[1]
			}
		}

		return a
	}

	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: level, AddSource: true, ReplaceAttr: replace},
	))
	slog.SetDefault(logger)
}
