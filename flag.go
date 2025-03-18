package catalystgo

import "flag"

var (
	configPath = flag.String("config", "./.catalystgo/config.yml", "CatalystGo config file")
)

func init() {
	flag.Parse()
}
