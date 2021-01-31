package generatespec

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func generateSpec() {
	proxyDir := "/home/ubuntu/go/src/github.com/jcollins-axway/apigee_to_swagger/apiproxy/"
	proxyName := "Petstore"
	// Open our proxyFile
	proxyFile, err := os.Open(proxyDir + "/" + proxyName + ".xml")
	if err != nil {
		fmt.Println(err)
	}
	byteValue, _ := ioutil.ReadAll(proxyFile)
	proxyFile.Close()

	// Unmarshal the proxy details
	var proxy apiProxy
	xml.Unmarshal(byteValue, &proxy)

	// Create the spec object
	spec := openapi3.Swagger{
		OpenAPI: "3.1.0",
		Info: &openapi3.Info{
			Title:       proxy.DisplayName,
			Description: proxy.Description,
			Version:     proxy.Version.Major + "." + proxy.Version.Minor,
		},
		Paths: openapi3.Paths{},
	}

	// Load the endpoint files
	for _, endpointFilename := range proxy.ProxyEndpoints.ProxyEndpoint {
		endpointFile, err := os.Open(proxyDir + "/proxies/" + endpointFilename + ".xml")
		if err != nil {
			fmt.Println(err)
		}

		byteValue, _ := ioutil.ReadAll(endpointFile)
		endpointFile.Close()
		generateEndpoints(&spec, byteValue)
	}

	swaggerBytes, _ := json.Marshal(spec)

	fmt.Println(string(swaggerBytes))

	err = spec.Validate(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}
}

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
