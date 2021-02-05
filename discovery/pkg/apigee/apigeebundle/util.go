package apigeebundle

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

//generateEndpoints - takes an APIGEE endpoints file and adds all endpoints to the spec
func generateEndpoints(spec *openapi3.Swagger, proxy proxyEndpoint) {
	// Unmarshal the proxy details
	for _, flow := range proxy.Flows.Flow {
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
