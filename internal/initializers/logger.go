package initializers

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	k8szap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	LOG_LEVEL_ENV = "LOG_LEVEL"
)

func NewLogger(devLogging bool, dryrun bool) (*logr.Logger, error) {
	logLevel := strings.ToLower(os.Getenv(LOG_LEVEL_ENV))

	var level zap.AtomicLevel
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		fmt.Println("Failed to parse Log Level , using info as default log level")
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	zapFields := []zap.Option{
		zap.Fields(
			zap.String("dryrun", strconv.FormatBool(dryrun)),
		),
	}

	opts := &k8szap.Options{
		Development: devLogging,
		Level:       level,
		ZapOpts:     zapFields,
	}

	opts.BindFlags(flag.CommandLine)

	logger := k8szap.New(k8szap.UseFlagOptions(opts))

	return &logger, nil
}
