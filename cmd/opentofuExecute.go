package cmd

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/opentofu"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type opentofuExecuteUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
}

type opentofuExecuteUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newOpentofuExecuteUtils() opentofuExecuteUtils {
	utils := opentofuExecuteUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func opentofuExecute(config opentofuExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *opentofuExecuteCommonPipelineEnvironment) {
	utils := newOpentofuExecuteUtils()

	err := runOpentofuExecute(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runOpentofuExecute(config *opentofuExecuteOptions, telemetryData *telemetry.CustomData, utils opentofuExecuteUtils, commonPipelineEnvironment *opentofuExecuteCommonPipelineEnvironment) error {
	if len(config.CliConfigFile) > 0 {
		utils.AppendEnv([]string{fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", config.CliConfigFile)})
	}

	if len(config.Workspace) > 0 {
		utils.AppendEnv([]string{fmt.Sprintf("TF_WORKSPACE=%s", config.Workspace)})
	}

	args := []string{}

	if slices.Contains([]string{"apply", "destroy"}, config.Command) {
		args = append(args, "-auto-approve")
	}
	if slices.Contains([]string{"apply", "plan"}, config.Command) && config.OpentofuSecrets != "" {
		args = append(args, fmt.Sprintf("-var-file=%s", config.OpentofuSecrets))
	}
	if slices.Contains([]string{"init", "validate", "plan", "apply", "destroy"}, config.Command) {
		args = append(args, "-no-color")
	}
	if config.AdditionalArgs != nil {
		args = append(args, config.AdditionalArgs...)
	}

	if config.Init {
		err := runOpentofu(utils, "init", []string{"-no-color"}, config.GlobalOptions)

		if err != nil {
			return err
		}
	}
	err := runOpentofu(utils, config.Command, args, config.GlobalOptions)
	if err != nil {
		return err
	}

	var outputBuffer bytes.Buffer
	utils.Stdout(&outputBuffer)

	err = runOpentofu(utils, "output", []string{"-json"}, config.GlobalOptions)

	if err != nil {
		return err
	}

	commonPipelineEnvironment.custom.opentofuOutputs, err = opentofu.ReadOutputs(outputBuffer.String())

	return nil
}

func runOpentofu(utils opentofuExecuteUtils, command string, additionalArgs []string, globalOptions []string) error {
	args := []string{}

	if len(globalOptions) > 0 {
		args = append(args, globalOptions...)
	}

	args = append(args, command)

	if len(additionalArgs) > 0 {
		args = append(args, additionalArgs...)
	}

	return utils.RunExecutable("tofu", args...)
}
