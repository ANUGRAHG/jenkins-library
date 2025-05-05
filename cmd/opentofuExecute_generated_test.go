//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpentofuExecuteCommand(t *testing.T) {
	t.Parallel()

	testCmd := OpentofuExecuteCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "opentofuExecute", testCmd.Use, "command name incorrect")

}
