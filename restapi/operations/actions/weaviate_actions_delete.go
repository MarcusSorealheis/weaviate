/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 - 2018 Weaviate. All rights reserved.
 * LICENSE: https://github.com/creativesoftwarefdn/weaviate/blob/develop/LICENSE.md
 * AUTHOR: Bob van Luijt (bob@kub.design)
 * See www.creativesoftwarefdn.org for details
 * Contact: @CreativeSofwFdn / bob@kub.design
 */
// Code generated by go-swagger; DO NOT EDIT.

package actions

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	context "golang.org/x/net/context"

	middleware "github.com/go-openapi/runtime/middleware"
)

// WeaviateActionsDeleteHandlerFunc turns a function with the right signature into a weaviate actions delete handler
type WeaviateActionsDeleteHandlerFunc func(context.Context, WeaviateActionsDeleteParams) middleware.Responder

// Handle executing the request and returning a response
func (fn WeaviateActionsDeleteHandlerFunc) Handle(ctx context.Context, params WeaviateActionsDeleteParams) middleware.Responder {
	return fn(ctx, params)
}

// WeaviateActionsDeleteHandler interface for that can handle valid weaviate actions delete params
type WeaviateActionsDeleteHandler interface {
	Handle(context.Context, WeaviateActionsDeleteParams) middleware.Responder
}

// NewWeaviateActionsDelete creates a new http.Handler for the weaviate actions delete operation
func NewWeaviateActionsDelete(ctx *middleware.Context, handler WeaviateActionsDeleteHandler) *WeaviateActionsDelete {
	return &WeaviateActionsDelete{Context: ctx, Handler: handler}
}

/*WeaviateActionsDelete swagger:route DELETE /actions/{actionId} actions weaviateActionsDelete

Delete an Action based on its UUID related to this key.

Deletes an Action from the system.

*/
type WeaviateActionsDelete struct {
	Context *middleware.Context
	Handler WeaviateActionsDeleteHandler
}

func (o *WeaviateActionsDelete) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewWeaviateActionsDeleteParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(r.Context(), Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
