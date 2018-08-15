package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/test-structure"
)

func TestDemo(t *testing.T) {
	t.Parallel()

	dir := "../environments/demo_env"

	defer test_structure.RunTestStage(t, "teardown", func() {
		options := test_structure.LoadTerraformOptions(t, dir)
		terraform.Destroy(t, options)

		keyPair := test_structure.LoadEc2KeyPair(t, dir)
		aws.DeleteEC2KeyPair(t, keyPair)
	})

	test_structure.RunTestStage(t, "setup", func() {
		options, keyPair := configureTerraformOptions(t, dir)

		test_structure.SaveTerraformOptions(t, dir, options)
		test_structure.SaveEc2KeyPair(t, dir, keyPair)

		terraform.InitAndApply(t, options)
	})

	test_structure.RunTestStage(t, "validate", func() {
		options := test_structure.LoadTerraformOptions(t, dir)
		keyPair := test_structure.LoadEc2KeyPair(t, dir)

		// Get instance IP from Terraform output.
		instanceIP := terraform.Output(t, options, "instance_public_ip")

		host := ssh.Host{
			Hostname:    instanceIP,
			SshKeyPair:  keyPair.KeyPair,
			SshUserName: "ec2-user",
		}
		maxRetries := 5
		retryInterval := 5 * time.Second
		description := fmt.Sprintf("Running command over SSH on host %s", instanceIP)
		expectedResult := "0"
		command := "curl -s google.com > /dev/null; echo $?"

		// Run command on host over SSH.
		retry.DoWithRetry(t, description, maxRetries, retryInterval, func() (string, error) {
			actualResult, err := ssh.CheckSshCommandE(t, host, command)

			if err != nil {
				return "", err
			}

			if strings.TrimSpace(actualResult) != expectedResult {
				return "", fmt.Errorf(
					"Expected command to return '%s' but got '%s'", expectedResult, actualResult,
				)
			}

			return "", nil
		})
	})
}

func configureTerraformOptions(t *testing.T, exampleFolder string) (*terraform.Options, *aws.Ec2Keypair) {
	region := "eu-central-1"
	// Create an EC2 KeyPair that we can use for SSH access
	keyPairName := "terratest-demo"
	keyPair := aws.CreateAndImportEC2KeyPair(t, region, keyPairName)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"region":       region,
			"cidr_block":   "172.16.0.0/16",
			"vpc_name":     "terratest-demo",
			"ssh_key_name": keyPairName,
		},
	}

	return terraformOptions, keyPair
}
