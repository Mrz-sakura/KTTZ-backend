package music

import (
	"app-bff/app/service/music"
	"app-bff/route"
)

func init() {
	route.Register(&music.Search{})
}
