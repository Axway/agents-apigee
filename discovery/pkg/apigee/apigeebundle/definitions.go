package apigeebundle

import (
	"encoding/xml"
	"regexp"
	"strings"

	"github.com/oriser/regroup"
)

// APIGEE XML structures
// APIProxy - APIGEE API Proxy definition
type APIProxy struct {
	Basepaths       string               `xml:"Basepaths"`
	Version         configurationVersion `xml:"ConfigurationVersion"`
	CreatedAt       int                  `xml:"CreatedAt"`
	CreatedBy       string               `xml:"CreatedBy"`
	Description     string               `xml:"Description"`
	DisplayName     string               `xml:"DisplayName"`
	LastModifiedAt  int                  `xml:"LastModifiedAt"`
	LastModifiedBy  string               `xml:"LastModifiedBy"`
	ManifestVersion string               `xml:"ManifestVersion"`
	Policies        policies             `xml:"Policies"`
	ProxyEndpoints  proxyEndpoints       `xml:"ProxyEndpoints"`
	Resources       string               `xml:"Resources"`
	Spec            string               `xml:"Spec"`
	TargetServers   string               `xml:"TargetServers"`
	TargetEndpoints string               `xml:"TargetEndpoints"`
}

// configurationVersion - APIGEE API Proxy version
type configurationVersion struct {
	Major string `xml:"majorVersion,attr"`
	Minor string `xml:"minorVersion,attr"`
}

// policies - APIGEE API Proxy Policy filenames
type policies struct {
	Policies []string `xml:"Policy"`
}

// proxyEndpoints - APIGEE API Proxy Endpoint filenames
type proxyEndpoints struct {
	ProxyEndpoint []string `xml:"ProxyEndpoint"`
}

// proxyEndpoint - APIGEE Proxy Endpoint file structure
type proxyEndpoint struct {
	Name    string  `xml:"name,attr"`
	PreFlow preFlow `xml:"PreFlow"`
	Flows   flows   `xml:"Flows"`
}

// preFlow - APIGEE Proxy endpoint preflow
type preFlow struct {
	Name     string   `xml:"name,attr"`
	Request  request  `xml:"Request"`
	Response response `xml:"Response"`
}

// request - APIGEE Proxy request steps
type request struct {
	Step []step `xml:"Step"`
}

// response - APIGEE Proxy response steps
type response struct {
	Step []step `xml:"Step"`
}

// step - APIGEE Proxy step names
type step struct {
	Name string `xml:"Name"`
}

// flows - APIGEE Proxy flows
type flows struct {
	Flow []flow `xml:"Flow"`
}

// flow - APIGEE proxy flow
type flow struct {
	Name        string     `xml:"name,attr"`
	Description string     `xml:"Description"`
	Request     request    `xml:"Request"`
	Response    response   `xml:"Response"`
	Conditions  conditions `xml:"Condition"`
}

// conditions - array of conditions
type conditions struct {
	Condition       []condition
	ConditionString string
}

// condition - parsed representation of APIGEE condition
type condition struct {
	Variable string
	Operator string
	Value    string
}

//UnmarshalXML - custom XML unmarshall to parse APIGEE flow conditions
func (c *conditions) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var el string

	if err := d.DecodeElement(&el, &start); err != nil {
		return err
	}

	if el == "null" {
		return nil
	}
	c.ConditionString = el

	// Split all conditions
	re := regexp.MustCompile("\\).*\\(")
	conds := re.Split(el, -1)

	// iterate conditions
	for _, cond := range conds {
		r := regroup.MustCompile("^\\({0,1}(?P<var>[a-z\\.]*) (?P<op>.*) (?P<val>.*\")\\){0,1}$")

		matches, err := r.Groups(cond)
		if err != nil {
			return err
		}

		// add each condition to the array
		c.Condition = append(c.Condition, condition{
			Variable: strings.Trim(matches["var"], "\""),
			Operator: strings.Trim(matches["op"], "\""),
			Value:    strings.Trim(matches["val"], "\""),
		})
	}

	return nil
}
