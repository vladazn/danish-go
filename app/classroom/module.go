package classroom

import "go.uber.org/fx"

var Module = fx.Module(
	"classroom",
	fx.Provide(
		NewFirebaseStore,
		NewDictionary,
		NewWordPool,
		NewSetService,
	),
)
