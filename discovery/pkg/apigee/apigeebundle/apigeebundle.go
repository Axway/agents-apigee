package apigeebundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

//APIGEEBundle -
type APIGEEBundle struct {
	Name            string
	ProxyDefinition APIProxy
	Proxies         map[string]proxyEndpoint
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
		default:
			log.Warnf("Skipped parsing %s", zipFile.Name)
		}
	}
}

//Generate - generate a spec file
func (a *APIGEEBundle) Generate(name, description, version string) []byte {
	// Create the serves array
	servers := []*openapi3.Server{}
	for _, url := range a.URLs {
		servers = append(servers, &openapi3.Server{
			URL: url,
		})
	}

	// data is the byte array of the zip archive
	spec := openapi3.Swagger{
		OpenAPI: "3.0.1",
		Info: &openapi3.Info{
			Title:       name,
			Description: description,
			Version:     version,
		},
		Paths:   openapi3.Paths{},
		Servers: servers,
	}

	for _, proxy := range a.Proxies {
		generateEndpoints(&spec, proxy)
	}

	specBytes, _ := json.Marshal(spec)
	return specBytes
}
