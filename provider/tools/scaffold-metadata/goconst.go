// Copyright 2016-2026, Pulumi Corporation.

package main

// Repeated operation-slot, module, and field-name literals, extracted so
// goconst stops flagging them.
const (
	opCreate     = "create"
	opRead       = "read"
	opUpdate     = "update"
	opDelete     = "delete"
	authModule   = "auth"
	nameFieldKey = "name"
)
