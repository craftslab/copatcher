package utils

import (
	"bufio"
	"io"

	log "github.com/sirupsen/logrus"
)

func LogPipe(pipe io.ReadCloser, level log.Level) {
	defer func(p io.ReadCloser) {
		_ = p.Close()
	}(pipe)

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		log.StandardLogger().Log(level, scanner.Text())
	}
}
