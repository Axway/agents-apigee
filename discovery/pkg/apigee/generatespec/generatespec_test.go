package generatespec

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestGenerateEndpointsGet(t *testing.T) {
	// Import test file
	endpointFile, _ := os.Open("./testdata/get_endpoint_passthrough.xml")

	// Run test
	spec := openapi3.Swagger{}
	byteValue, _ := ioutil.ReadAll(endpointFile)
	endpointFile.Close()
	GenerateEndpoints(&spec, byteValue)

	// Validations
	assert.Equal(t, 1, len(spec.Paths), "Only 1 path was expected")
	pathItem := spec.Paths.Find("/pet/{petId}")
	assert.NotNil(t, pathItem, "The /pet/{petId} was not found")
	assert.NotNil(t, pathItem.Get, "Expected the Get operation to be defined")
	assert.NotNil(t, pathItem.Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "petId"), "Expected petId parameter defined for path")
}
