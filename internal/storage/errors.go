package storage

import "errors"

var ErrRights = errors.New("insufficient rights to perform the action")
var ErrNoTender = errors.New("no such tender")
var ErrNoBid = errors.New("no such bid")
var ErrNoVersion = errors.New("no such version")
var ErrNoReviews = errors.New("no such review")
var ErrNoUser = errors.New("no such user")
