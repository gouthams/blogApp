/*
 * Simple blogging APIs
 *
 * This is a simple blogging API
 *
 * API version: 1.0.0
 * Contact: gouthams.ku@gmail.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package restimpl

import (
	"time"
)

type BlogPost struct {
	Id string `json:"id,omitempty"`

	UserId string `json:"userId" binding:"required"`

	Topic string `json:"topic" binding:"required"`

	Content string `json:"content" binding:"required"`

	LastModifiedDate time.Time `json:"lastModifiedDate,omitempty"`
}
