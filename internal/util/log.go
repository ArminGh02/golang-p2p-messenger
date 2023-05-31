package util

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Logger(cmd *cobra.Command) *logrus.Logger {
	logger := logrus.New()
	logger.Out = cmd.OutOrStdout()
	return logger
}
