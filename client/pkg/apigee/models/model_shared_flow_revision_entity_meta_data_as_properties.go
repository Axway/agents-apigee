/*
 * Shared flows and flow hooks API
 *
 * Manage shared flows and flow hooks. For more information, see: * <a href=\"https://docs.apigee.com/api-platform/fundamentals/shared-flows\">Reusable shared flows</a> * <a href=\"https://docs.apigee.com/api-platform/fundamentals/flow-hooks\">Attaching a shared flow using a flow hook</a>.
 *
 * API version: 1.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package models

// SharedFlowRevisionEntityMetaDataAsProperties Kev-value map of metadata describing the shared flow revision.
type SharedFlowRevisionEntityMetaDataAsProperties struct {
	// Type of bundle. Set to `zip`.
	BundleType string `json:"bundle_type,omitempty"`
	// Time when the shared flow revision was created in milliseconds since epoch.
	CreatedAt string `json:"createdAt,omitempty"`
	// Email address of developer that created the shared flow.
	CreatedBy string `json:"createdBy,omitempty"`
	// Time when the shared flow version was last modified in milliseconds since epoch.
	LastModifiedAt string `json:"lastModifiedAt,omitempty"`
	// Email address of developer that last modified the shared flow.
	LastModifiedBy string `json:"lastModifiedBy,omitempty"`
	// Set to `null`.
	SubType string `json:"subType,omitempty"`
}
