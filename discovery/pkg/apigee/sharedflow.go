package apigee

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/markbates/pkger"
)

const (
	sharedFlow = "amplify-central-logging"
)

// addSharedFlow - checks to see if the logging flow has been added and adds if it hasn't
func (a *GatewayClient) addSharedFlow() {
	_, err := a.getSharedFlow(sharedFlow)
	if err != nil {
		return
	}

	// flow does not exist, add it
	data, _ := a.createSharedFlowZip()

	// upload the flow to apigee
	a.createSharedFlow(data, sharedFlow)

	// TODO - get the posted shared flow latest revision for deployment call

	// deploy the shared flow to all envs and create the hook
	for env := range a.envToURLs {
		a.deploySharedFlow(env, sharedFlow, "1")
		a.publishSharedFlowToEnvironment(env, sharedFlow)
	}
}

// createSharedFlowZip - creates the shared flow bundle from the template files
func (a *GatewayClient) createSharedFlowZip() ([]byte, error) {

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
		if zipPath == "/sharedflowbundle/policies/amplify-central-logging.xml" {
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
func (a *GatewayClient) updateSharedFlowPolicy(templateBytes []byte) ([]byte, error) {
	// Will hold the byte array with the file contents
	data := bytes.Buffer{}

	tmpl, _ := template.New("policy").Parse(string(templateBytes))

	/// TODO - use loggly config here
	type TemplateData struct {
		APIToken string
	}

	templateData := TemplateData{"7fa9ef24-0fba-407f-af34-103ce8872124"}
	err := tmpl.Execute(&data, templateData)

	return data.Bytes(), err
}
