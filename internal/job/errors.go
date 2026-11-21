package job

import "errors"

var ErrStateConflicted = errors.New("job not in expected state")
