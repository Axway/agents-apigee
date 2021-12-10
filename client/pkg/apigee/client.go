package apigee

import (
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/jobs"

	"github.com/Axway/agents-apigee/client/pkg/config"
)

// ApigeeClient - Represents the Gateway client
type ApigeeClient struct {
	cfg          *config.ApigeeConfig
	apiClient    coreapi.Client
	accessToken  string
	pollInterval time.Duration
	envToURLs    map[string][]string
	isReady      bool
}

// NewClient - Creates a new Gateway Client
func NewClient(apigeeCfg *config.ApigeeConfig) (*ApigeeClient, error) {
	client := &ApigeeClient{
		apiClient:    coreapi.NewClient(nil, ""),
		cfg:          apigeeCfg,
		pollInterval: apigeeCfg.GetPollInterval(),
		envToURLs:    make(map[string][]string),
		isReady:      false,
	}

	// create the auth job and register it
	authentication := newAuthJob(client.apiClient, apigeeCfg.Auth.GetUsername(), apigeeCfg.Auth.GetPassword(), client.setAccessToken)
	jobs.RegisterIntervalJobWithName(authentication, 10*time.Minute, "APIGEE Auth Token")

	return client, nil
}

func (a *ApigeeClient) setAccessToken(token string) {
	a.accessToken = token
	a.isReady = true
}

// AddSharedFlow - checks to see if the logging flow has been added and adds if it hasn't
// func (a *ApigeeClient) AddSharedFlow(sharedFlow string) {
// 	log.Debugf("Checking for shared flow")
// 	_, err := a.GetSharedFlow(sharedFlow)
// 	if err == nil {
// 		return
// 	}
// 	log.Debugf("Shared flow not found, deploying it now")

// 	// flow does not exist, add it
// 	data, _ := a.createSharedFlowZip()

// 	// upload the flow to apigee
// 	log.Debugf("Deploy shared flow to apigee")
// 	err = a.CreateSharedFlow(data, sharedFlow)
// 	if err != nil {
// 		log.Errorf("Error hit deploying the shared flow: %v", err)
// 	}
// 	// TODO - get the posted shared flow latest revision for deployment call

// 	// deploy the shared flow to all envs and create the hook
// 	for env := range a.envToURLs {
// 		log.Debugf("Deploy flow hook to %s", env)
// 		err = a.DeploySharedFlow(env, sharedFlow, "1")
// 		if err != nil {
// 			log.Errorf("Error hit deploying the shared flow revision to the %s env: %v", env, err)
// 		}
// 		err = a.PublishSharedFlowToEnvironment(env, sharedFlow)
// 		if err != nil {
// 			log.Errorf("Error hit publising the shared flow to the %s env: %v", env, err)
// 		}
// 	}
// }

// // createSharedFlowZip - creates the shared flow bundle from the template files
// func (a *ApigeeClient) createSharedFlowZip() ([]byte, error) {
// 	log.Debugf("Creating archive for shared flow")

// 	newZipFile := new(bytes.Buffer)
// 	zipWriter := zip.NewWriter(newZipFile)

// 	err := pkger.Walk("/sharedflowbundle", func(path string, info os.FileInfo, err error) error {
// 		// skip directories
// 		if info.IsDir() {
// 			return nil
// 		}

// 		// Open the file
// 		f, err := pkger.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer f.Close()

// 		// Create the file in the zip
// 		zipPath := strings.Split(path, ":/")[1]
// 		zipFile, err := zipWriter.Create(zipPath)
// 		if err != nil {
// 			return err
// 		}

// 		// read the file contents
// 		filedata, err := ioutil.ReadAll(f)
// 		if err != nil {
// 			return err
// 		}

// 		// if this file is the policy, we have to update it
// 		if zipPath == "sharedflowbundle/policies/amplify-central-logging.xml" {
// 			filedata, err = a.updateSharedFlowPolicy(filedata)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		// write the file to the zip file
// 		_, err = zipFile.Write(filedata)
// 		if err != nil {
// 			return err
// 		}

// 		return nil
// 	})

// 	zipWriter.Close()
// 	return newZipFile.Bytes(), err
// }

// //updateSharedFlowPolicy - updates the shared flow policy file with the appropriate loggly settings
// func (a *ApigeeClient) updateSharedFlowPolicy(templateBytes []byte) ([]byte, error) {
// 	// Will hold the byte array with the file contents
// 	data := bytes.Buffer{}

// 	tmpl, _ := template.New("policy").Parse(string(templateBytes))
// 	err := tmpl.Execute(&data, a.cfg.GetLoggly())

// 	return data.Bytes(), err
// }

//GetConfig - return the apigee client config
func (a *ApigeeClient) GetConfig() *config.ApigeeConfig {
	return a.cfg
}

//IsReady - returns true when the apigee client authenticates
func (a *ApigeeClient) IsReady() bool {
	return a.isReady
}
