/*
 * Developers API
 *
 * Developers must register with an organization on Apigee Edge. After they are registered, developers register their apps, choose the APIs they want to use, and receive the unique API credentials (consumer keys and secrets) needed to access your APIs.
 *
 * API version: 1.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package models

// DeveloperRequest Developer request.
type DeveloperRequest struct {
	// List of attributes that can be used to extend the default developer profile. With Apigee Edge for Public Cloud, the custom attribute limit is 18.
	Attributes []Attribute `json:"attributes,omitempty"`
	// Email address of the developer. This value is used to uniquely identify the developer in Apigee Edge.
	Email string `json:"email"`
	// First name of the developer.
	FirstName string `json:"firstName"`
	// Last name of the developer.
	LastName string `json:"lastName"`
	// Username. Not used by Apigee.
	UserName string `json:"userName"`
}
