//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2023 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

// Code generated by go-swagger; DO NOT EDIT.

package meta

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewMetaGetParams creates a new MetaGetParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewMetaGetParams() *MetaGetParams {
	return &MetaGetParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewMetaGetParamsWithTimeout creates a new MetaGetParams object
// with the ability to set a timeout on a request.
func NewMetaGetParamsWithTimeout(timeout time.Duration) *MetaGetParams {
	return &MetaGetParams{
		timeout: timeout,
	}
}

// NewMetaGetParamsWithContext creates a new MetaGetParams object
// with the ability to set a context for a request.
func NewMetaGetParamsWithContext(ctx context.Context) *MetaGetParams {
	return &MetaGetParams{
		Context: ctx,
	}
}

// NewMetaGetParamsWithHTTPClient creates a new MetaGetParams object
// with the ability to set a custom HTTPClient for a request.
func NewMetaGetParamsWithHTTPClient(client *http.Client) *MetaGetParams {
	return &MetaGetParams{
		HTTPClient: client,
	}
}

/*
MetaGetParams contains all the parameters to send to the API endpoint

	for the meta get operation.

	Typically these are written to a http.Request.
*/
type MetaGetParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the meta get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *MetaGetParams) WithDefaults() *MetaGetParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the meta get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *MetaGetParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the meta get params
func (o *MetaGetParams) WithTimeout(timeout time.Duration) *MetaGetParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the meta get params
func (o *MetaGetParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the meta get params
func (o *MetaGetParams) WithContext(ctx context.Context) *MetaGetParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the meta get params
func (o *MetaGetParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the meta get params
func (o *MetaGetParams) WithHTTPClient(client *http.Client) *MetaGetParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the meta get params
func (o *MetaGetParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *MetaGetParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
