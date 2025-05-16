package asyncProvider

import (
	"github.com/google/wire"
)

var AsyncWireSet = wire.NewSet(
	NewAsyncRunnable,
)
