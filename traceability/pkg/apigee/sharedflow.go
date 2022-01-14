package apigee

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	logglycfg "github.com/Axway/agents-apigee/traceability/pkg/config"
	"github.com/markbates/pkger"
)

const (
	sharedFlow = "amplify-central-logging"
)

type registerSharedFlowJob struct {
	jobs.Job
	apigeeClient *apigee.ApigeeClient
	logglyCfg    *logglycfg.LogglyConfig
	envToURLs    map[string][]string
}

func newRegisterSharedFlowJob(apigeeClient *apigee.ApigeeClient, logglyCfg *logglycfg.LogglyConfig) *registerSharedFlowJob {
	return &registerSharedFlowJob{
		apigeeClient: apigeeClient,
		envToURLs:    make(map[string][]string),
		logglyCfg:    logglyCfg,
	}
}

func (j *registerSharedFlowJob) Ready() bool {
	return j.apigeeClient.IsReady()
}

func (j *registerSharedFlowJob) Status() error {
	return nil
}

func (j *registerSharedFlowJob) Execute() error {
	j.addSharedFlow()
	return nil
}

// addSharedFlow - checks to see if the logging flow has been added and adds if it hasn't
func (j *registerSharedFlowJob) addSharedFlow() {
	log.Debugf("Checking for shared flow")
	_, err := j.apigeeClient.GetSharedFlow(sharedFlow)
	if err == nil {
		return
	}
	log.Debugf("Shared flow not found, deploying it now")

	// flow does not exist, add it
	data, _ := j.createSharedFlowZip()

	// upload the flow to apigee
	log.Debugf("Deploy shared flow to apigee")
	err = j.apigeeClient.CreateSharedFlow(data, sharedFlow)
	if err != nil {
		log.Errorf("Error hit deploying the shared flow: %v", err)
	}
	// TODO - get the posted shared flow latest revision for deployment call

	// deploy the shared flow to all envs and create the hook
	for env := range j.envToURLs {
		log.Debugf("Deploy flow hook to %s", env)
		err = j.apigeeClient.DeploySharedFlow(env, sharedFlow, "1")
		if err != nil {
			log.Errorf("Error hit deploying the shared flow revision to the %s env: %v", env, err)
		}
		err = j.apigeeClient.PublishSharedFlowToEnvironment(env, sharedFlow)
		if err != nil {
			log.Errorf("Error hit publising the shared flow to the %s env: %v", env, err)
		}
	}
}

// createSharedFlowZip - creates the shared flow bundle from the template files
func (j *registerSharedFlowJob) createSharedFlowZip() ([]byte, error) {
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
			filedata, err = j.updateSharedFlowPolicy(filedata)
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
func (j *registerSharedFlowJob) updateSharedFlowPolicy(templateBytes []byte) ([]byte, error) {
	// Will hold the byte array with the file contents
	data := bytes.Buffer{}

	tmpl, _ := template.New("policy").Parse(string(templateBytes))
	err := tmpl.Execute(&data, j.logglyCfg)

	return data.Bytes(), err
}
