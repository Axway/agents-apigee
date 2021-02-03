package apigeebundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/discovery/pkg/apigee/models"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

type APIGEEBundle struct {
	Name            string
	ZipContents     []byte
	ProxyDefinition APIProxy
	Policies        interface{}
	Proxies         map[string][]proxyEndpoint
	Resources       interface{}
}

func NewAPIGEEBundle(data []byte, proxyName string) *APIGEEBundle {
	bundle := APIGEEBundle{
		Name:            proxyName,
		ZipContents:     data,
		ProxyDefinition: APIProxy{},
		Proxies:         map[string][]proxyEndpoint{},
	}

	bundle.parseAll()
	return &bundle
}

func (a *APIGEEBundle) parseAll() {
	zipReader, err := zip.NewReader(bytes.NewReader(a.ZipContents), int64(len(a.ZipContents)))
	if err != nil {
		log.Error(err)
	}

	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		// we only care about the files in proxies
		switch {
		case strings.HasPrefix(zipFile.Name, "apiproxy/"+a.Name+".xml"):
			// Proxy definition file
			fileBytes, err := util.ReadZipFile(zipFile)
			if err != nil {
				log.Error(err)
			}
			xml.Unmarshal(fileBytes, &a.ProxyDefinition)
		case strings.HasPrefix(zipFile.Name, "apiproxy/proxies/") && strings.HasSuffix(zipFile.Name, ".xml"):
			var endpoint proxyEndpoint
			// Proxy proxies file
			fileBytes, err := util.ReadZipFile(zipFile)
			if err != nil {
				log.Error(err)
				continue
			}
			xml.Unmarshal(fileBytes, &endpoint)
			a.Proxies[zipFile.Name] = append(a.Proxies[zipFile.Name], endpoint)
		default:
			log.Warnf("Skipped parsing %s", zipFile.Name)
		}
	}
}

//Generate - generate a spec file
func (a *APIGEEBundle) Generate(revision models.ApiProxyRevision, apigeeURLs []string, basepath string) []byte {
	// Create the serves array
	servers := []*openapi3.Server{}
	for _, apigeeURL := range apigeeURLs {
		servers = append(servers, &openapi3.Server{
			URL: apigeeURL + basepath,
		})
	}

	// data is the byte array of the zip archive
	spec := openapi3.Swagger{
		OpenAPI: "3.0.1",
		Info: &openapi3.Info{
			Title:       revision.Name,
			Description: revision.Description,
			Version:     fmt.Sprintf("%d.%d", revision.ConfigurationVersion.MajorVersion, revision.ConfigurationVersion.MinorVersion),
		},
		Paths:   openapi3.Paths{},
		Servers: servers,
	}

	zipReader, err := zip.NewReader(bytes.NewReader(a.ZipContents), int64(len(a.ZipContents)))
	if err != nil {
		log.Error(err)
	}

	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		// we only care about the files in proxies
		if strings.HasPrefix(zipFile.Name, "apiproxy/proxies/") && strings.HasSuffix(zipFile.Name, ".xml") {
			fileBytes, err := util.ReadZipFile(zipFile)
			if err != nil {
				log.Error(err)
				continue
			}
			generateEndpoints(&spec, fileBytes)
		}
	}

	specBytes, _ := json.Marshal(spec)
	return specBytes
}

//GenerateEndpoints - takes an APIGEE endpoints file and adds all endpoints to the spec
func generateEndpoints(spec *openapi3.Swagger, filedata []byte) {
	// Unmarshal the proxy details
	var endpoint proxyEndpoint
	xml.Unmarshal(filedata, &endpoint)

	for _, flow := range endpoint.Flows.Flow {
		var verb, urlPath string
		operation := openapi3.Operation{
			OperationID: flow.Name,
			Description: flow.Description,
			Summary:     flow.Description,
		}
		for _, condition := range flow.Conditions.Condition {
			if condition.Variable == "proxy.pathsuffix" && condition.Operator == "MatchesPath" {
				urlPath = condition.Value
				// Split path
				pathComponents := strings.Split(urlPath, "/")
				for i, pathComponent := range pathComponents {
					if pathComponent == "*" {
						paramName := pathComponents[i-1] + "Id"
						// This is a * part of the url, change it to a variable name based on previous component
						pathComponents[i] = "{" + paramName + "}"
						operation.AddParameter(&openapi3.Parameter{
							In:       openapi3.ParameterInPath,
							Name:     paramName,
							Schema:   openapi3.NewSchemaRef("string", openapi3.NewStringSchema()),
							Required: true,
						})
					}
				}
				urlPath = strings.Join(pathComponents, "/")
			} else if condition.Variable == "request.verb" && (condition.Operator == "=" || condition.Operator == "equal") {
				verb = condition.Value

			}
		}
		operation.AddResponse(200, &openapi3.Response{
			Description: &flow.Description,
		})

		spec.AddOperation(urlPath, verb, &operation)
	}
}
