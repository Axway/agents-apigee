package apigee

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	corecfg "github.com/Axway/agent-sdk/pkg/config"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/markbates/pkger"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/config"

	logglycfg "github.com/Axway/agents-apigee/traceability/pkg/config"
)

// AgentConfig - represents the config for agent
type AgentConfig struct {
	CentralCfg corecfg.CentralConfig   `config:"central"`
	ApigeeCfg  *config.ApigeeConfig    `config:"apigee"`
	LogglyCfg  *logglycfg.LogglyConfig `config:"loggly"`
}

// Agent - Represents the Gateway client
type Agent struct {
	cfg          *AgentConfig
	apigeeClient *apigee.ApigeeClient
	envToURLs    map[string][]string
}

// NewAgent - Creates a new Agent
func NewAgent(agentCfg *AgentConfig) (*Agent, error) {
	apigeeClient, err := apigee.NewClient(agentCfg.ApigeeCfg)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		apigeeClient: apigeeClient,
		cfg:          agentCfg,
		envToURLs:    make(map[string][]string),
	}

	return agent, nil
}

const (
	sharedFlow = "amplify-central-logging"
)

// addSharedFlow - checks to see if the logging flow has been added and adds if it hasn't
func (a *Agent) addSharedFlow() {
	log.Debugf("Checking for shared flow")
	_, err := a.apigeeClient.GetSharedFlow(sharedFlow)
	if err == nil {
		return
	}
	log.Debugf("Shared flow not found, deploying it now")

	// flow does not exist, add it
	data, _ := a.createSharedFlowZip()

	// upload the flow to apigee
	log.Debugf("Deploy shared flow to apigee")
	err = a.apigeeClient.CreateSharedFlow(data, sharedFlow)
	if err != nil {
		log.Errorf("Error hit deploying the shared flow: %v", err)
	}
	// TODO - get the posted shared flow latest revision for deployment call

	// deploy the shared flow to all envs and create the hook
	for env := range a.envToURLs {
		log.Debugf("Deploy flow hook to %s", env)
		err = a.apigeeClient.DeploySharedFlow(env, sharedFlow, "1")
		if err != nil {
			log.Errorf("Error hit deploying the shared flow revision to the %s env: %v", env, err)
		}
		err = a.apigeeClient.PublishSharedFlowToEnvironment(env, sharedFlow)
		if err != nil {
			log.Errorf("Error hit publising the shared flow to the %s env: %v", env, err)
		}
	}
}

// createSharedFlowZip - creates the shared flow bundle from the template files
func (a *Agent) createSharedFlowZip() ([]byte, error) {
	log.Debugf("Creating archive for shared flow")

	newZipFile := new(bytes.Buffer)
	zipWriter := zip.NewWriter(newZipFile)

	err := pkger.Walk("/sharedflowbundle", func(path string, info os.FileInfo, err error) error {
		// skip directories
		if info.IsDir() {
			return nil
		}

		// Open the file
		f, err := pkger.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// Create the file in the zip
		zipPath := strings.Split(path, ":/")[1]
		zipFile, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		// read the file contents
		filedata, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		// if this file is the policy, we have to update it
		if zipPath == "sharedflowbundle/policies/amplify-central-logging.xml" {
			filedata, err = a.updateSharedFlowPolicy(filedata)
			if err != nil {
				return err
			}
		}

		// write the file to the zip file
		_, err = zipFile.Write(filedata)
		if err != nil {
			return err
		}

		return nil
	})

	zipWriter.Close()
	return newZipFile.Bytes(), err
}

//updateSharedFlowPolicy - updates the shared flow policy file with the appropriate loggly settings
func (a *Agent) updateSharedFlowPolicy(templateBytes []byte) ([]byte, error) {
	// Will hold the byte array with the file contents
	data := bytes.Buffer{}

	tmpl, _ := template.New("policy").Parse(string(templateBytes))
	err := tmpl.Execute(&data, a.cfg.LogglyCfg)

	return data.Bytes(), err
}
