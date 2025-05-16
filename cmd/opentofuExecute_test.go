package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type opentofuExecuteMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newOpentofuExecuteTestsUtils() opentofuExecuteMockUtils {
	utils := opentofuExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunOpentofuExecute(t *testing.T) {
	t.Parallel()
	tt := []struct {
		opentofuExecuteOptions
		expectedArgs    []string
		expectedEnvVars []string
	}{
		{
			opentofuExecuteOptions{
				Command: "apply",
			}, []string{"apply", "-auto-approve", "-no-color"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command:         "apply",
				OpentofuSecrets: "/tmp/test",
			}, []string{"apply", "-auto-approve", "-var-file=/tmp/test", "-no-color"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command: "plan",
			}, []string{"plan", "-no-color"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command:         "plan",
				OpentofuSecrets: "/tmp/test",
			}, []string{"plan", "-var-file=/tmp/test", "-no-color"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command:         "plan",
				OpentofuSecrets: "/tmp/test",
				AdditionalArgs:  []string{"-arg1"},
			}, []string{"plan", "-var-file=/tmp/test", "-no-color", "-arg1"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command:         "apply",
				OpentofuSecrets: "/tmp/test",
				AdditionalArgs:  []string{"-arg1"},
			}, []string{"apply", "-auto-approve", "-var-file=/tmp/test", "-no-color", "-arg1"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command:         "apply",
				OpentofuSecrets: "/tmp/test",
				AdditionalArgs:  []string{"-arg1"},
				GlobalOptions:   []string{"-chdir=src"},
			}, []string{"-chdir=src", "apply", "-auto-approve", "-var-file=/tmp/test", "-no-color", "-arg1"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command: "apply",
				Init:    true,
			}, []string{"apply", "-auto-approve", "-no-color"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command:       "apply",
				GlobalOptions: []string{"-chdir=src"},
				Init:          true,
			}, []string{"-chdir=src", "apply", "-auto-approve", "-no-color"}, []string{},
		},
		{
			opentofuExecuteOptions{
				Command:       "apply",
				CliConfigFile: ".pipeline/.tofurc",
			}, []string{"apply", "-auto-approve", "-no-color"}, []string{"TF_CLI_CONFIG_FILE=.pipeline/.tofurc"},
		},
		{
			opentofuExecuteOptions{
				Command:   "plan",
				Workspace: "any-workspace",
			}, []string{"plan", "-no-color"}, []string{"TF_WORKSPACE=any-workspace"},
		},
	}

	for i, test := range tt {
		t.Run(fmt.Sprintf("That arguments are correct %d", i), func(t *testing.T) {
			t.Parallel()
			// init
			config := test.opentofuExecuteOptions
			utils := newOpentofuExecuteTestsUtils()
			utils.StdoutReturn = map[string]string{}
			utils.StdoutReturn["tofu output -json"] = "{}"
			utils.StdoutReturn["tofu -chdir=src output -json"] = "{}"

			runner := utils.ExecMockRunner

			// test
			err := runOpentofuExecute(&config, nil, utils, &opentofuExecuteCommonPipelineEnvironment{})

			// assert
			assert.NoError(t, err)

			if config.Init {
				assert.Equal(t, mock.ExecCall{Exec: "tofu", Params: append(config.GlobalOptions, "init", "-no-color")}, utils.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "tofu", Params: test.expectedArgs}, utils.Calls[1])
			} else {
				assert.Equal(t, mock.ExecCall{Exec: "tofu", Params: test.expectedArgs}, utils.Calls[0])
			}

			assert.Subset(t, runner.Env, test.expectedEnvVars)
		})
	}

	t.Run("Outputs get injected into CPE", func(t *testing.T) {
		t.Parallel()

		cpe := opentofuExecuteCommonPipelineEnvironment{}

		config := opentofuExecuteOptions{
			Command: "plan",
		}
		utils := newOpentofuExecuteTestsUtils()
		utils.StdoutReturn = map[string]string{}
		utils.StdoutReturn["tofu output -json"] = `{
			"sample_var": {
				"sensitive": true,
				"value": "a secret value",
				"type": "string"
			}
}
		`

		// test
		err := runOpentofuExecute(&config, nil, utils, &cpe)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, 1, len(cpe.custom.opentofuOutputs))
		assert.Equal(t, "a secret value", cpe.custom.opentofuOutputs["sample_var"])
	})
}
