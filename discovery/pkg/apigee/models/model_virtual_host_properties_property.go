/*
 * Virtual hosts API
 *
 * Manage virtual hosts. Virtual hosts let multiple domain names connect to the same host. A virtual host on Edge defines the domains and ports on which an API proxy is exposed, and, by extension, the URL that apps use to access an API proxy. A virtual host also defines whether the API proxy is accessed by using the HTTP protocol, or by the encrypted HTTPS protocol.
 *
 * API version: 1.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package models

// VirtualHostPropertiesProperty struct for VirtualHostPropertiesProperty
type VirtualHostPropertiesProperty struct {
	// Property name.
	Name string `json:"_name,omitempty"`
	// Property value.
	Text string `json:"__text,omitempty"`
}
