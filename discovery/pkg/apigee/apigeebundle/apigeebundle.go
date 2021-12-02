package apigeebundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tidwall/gjson"

	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	apiKeyName = "ApiKeyAuth"
)

//APIGEEBundle -
type APIGEEBundle struct {
	Name            string
	ProxyDefinition APIProxy
	Proxies         map[string]proxyEndpoint
	VerifyAPIKey    apiKeyPolicy
	URLs            []string
}

//NewAPIGEEBundle - instantiator of the APIGEEBundle, takes a bundle zip file and parses it to structures
func NewAPIGEEBundle(data []byte, proxyName string, apigeeURLs []string) *APIGEEBundle {
	bundle := APIGEEBundle{
		Name:            proxyName,
		ProxyDefinition: APIProxy{},
		Proxies:         map[string]proxyEndpoint{},
	}

	bundle.parseAll(data)
	bundle.setupURLs(apigeeURLs)
	return &bundle
}

//setupURLs - takes an array of base urls and adds the proxy's basepath to each
func (a *APIGEEBundle) setupURLs(apigeeURLs []string) {
	for _, apigeeURL := range apigeeURLs {
		a.URLs = append(a.URLs, apigeeURL+a.ProxyDefinition.Basepaths)
	}
}

//parseAll - takes the bundle byte array and iterates to fins all files ot build structures needed to create API Services
func (a *APIGEEBundle) parseAll(data []byte) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
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
			a.Proxies[zipFile.Name] = endpoint
		case strings.HasPrefix(zipFile.Name, "apiproxy/policies/verify-api-key.xml"):
			// APIKey policy definition
			fileBytes, err := util.ReadZipFile(zipFile)
			if err != nil {
				log.Error(err)
			}
			xml.Unmarshal(fileBytes, &a.VerifyAPIKey)
		default:
			log.Debugf("Skipped parsing %s", zipFile.Name)
		}
	}
}

//UpdateSpec - updates the URLS in the spec
func (a *APIGEEBundle) UpdateSpec(spec []byte) []byte {
	version := gjson.GetBytes(spec, "swagger").String()
	if version != "" {
		return a.updateOAS2Spec(spec)
	}
	return a.updateOAS3Spec(spec)
}

func (a *APIGEEBundle) updateOAS2Spec(spec []byte) []byte {
	//OAS2

	// Update the host
	oas2Spec := openapi2.T{}
	json.Unmarshal(spec, &oas2Spec)
	oas2Spec.Schemes = []string{"http"}
	for _, urlString := range a.URLs {
		urlDetails, _ := url.Parse(urlString)
		oas2Spec.Host = urlDetails.Host
		oas2Spec.BasePath = urlDetails.Path
		if urlDetails.Scheme == "https" {
			oas2Spec.Schemes = []string{urlDetails.Scheme}
		}
	}

	newSpec, _ := json.Marshal(oas2Spec)
	return newSpec
}

func (a *APIGEEBundle) updateOAS3Spec(spec []byte) []byte {
	//OAS3

	// Update the servers array
	servers := []*openapi3.Server{}
	for _, url := range a.URLs {
		servers = append(servers, &openapi3.Server{
			URL: url,
		})
	}
	oas3Spec := openapi3.T{}
	json.Unmarshal(spec, &oas3Spec)
	oas3Spec.Servers = servers

	newSpec, _ := json.Marshal(oas3Spec)
	return newSpec
}

//Generate - generate a spec file
func (a *APIGEEBundle) Generate(name, description, version string) []byte {
	// data is the byte array of the zip archive
	spec := openapi3.T{
		OpenAPI: "3.0.1",
		Info: &openapi3.Info{
			Title:       name,
			Description: description,
			Version:     version,
		},
		Paths: openapi3.Paths{},
	}

	// Update the security policy
	if a.VerifyAPIKey.Enabled == "true" {
		// Update the security scheme
		securityScheme := openapi3.SecurityScheme{
			Name: a.VerifyAPIKey.APIKey.Key,
			In:   a.VerifyAPIKey.APIKey.Location,
			Type: "apiKey",
		}
		securitySchemeRef := openapi3.SecuritySchemeRef{
			Value: &securityScheme,
		}
		spec.Components.SecuritySchemes = openapi3.SecuritySchemes{apiKeyName: &securitySchemeRef}

		// Add the scheme to security
		spec.Security = []openapi3.SecurityRequirement{
			{apiKeyName: []string{}},
		}
	}

	for _, proxy := range a.Proxies {
		generateEndpoints(&spec, proxy)
	}

	specBytes, _ := json.Marshal(spec)
	return specBytes
}
