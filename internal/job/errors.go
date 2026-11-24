package job

import "errors"

var ErrStateConflicted = errors.New("job not in expected state")
var ErrNotFound = errors.New("job not found")
var ErrKeyReuse = errors.New("idempotency key reused with a different request")
