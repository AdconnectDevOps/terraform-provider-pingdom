package main

import (
	"github.com/AdconnectDevOps/terraform-provider-pingdom/pingdom"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: pingdom.Provider,
	})
}
