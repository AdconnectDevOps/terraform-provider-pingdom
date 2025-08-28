package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/AdconnectDevOps/terraform-provider-pingdom/pingdom"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: pingdom.Provider,
	})
}
