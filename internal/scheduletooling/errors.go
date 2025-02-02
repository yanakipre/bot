package scheduletooling

import (
	"github.com/yanakipe/bot/internal/semerr"
)

func (j *InProcessJob) toSemanticErr(err error) *semerr.Error {
	var serr *semerr.Error
	if e := semerr.AsSemanticError(err); e == nil {
		serr = semerr.WrapWithInternal(err, "internal error")
	} else {
		serr = e
	}
	return serr
}
