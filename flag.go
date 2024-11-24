package catalystgo

import "flag"

var (
	configPath = flag.String("config", "./.catalystgo/config-local.yml", "CatalystGo config file")
)

func init() {
	flag.Parse()
}
