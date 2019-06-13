package dkvs

import "errors"

var ERROR_NOT_IMPLEMENTED = errors.New("not implemented")

var ERROR_KEY_NOT_FOUND = errors.New("key not found")
var ERROR_NOT_MASTER = errors.New("this node isn't the master")
