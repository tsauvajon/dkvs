package dkvs

import "errors"

var errorNotImplemented = errors.New("not implemented")

var errorKeyNotFound = errors.New("key not found")
var errorNotMaster = errors.New("this node isn't the master")
