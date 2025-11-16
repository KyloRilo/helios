package consul

import (
	"log"
	"os"

	"github.com/hashicorp/consul/api"
)

type ConsulController struct {
	client api.Client
	addr   string
	config *api.Config
}

func (c ConsulController) GetConfig() *api.Config {
	return c.config
}

func (c ConsulController) GetServices() {
	servcs, _, err := c.client.Catalog().Services(nil)
	if err != nil {
		log.Printf("Error querying services: %v", err)
	} else {
		log.Println("Services in catalog:")
		for servc := range servcs {
			log.Printf("- %s", servc)
		}
	}
}

func NewConsulController() *ConsulController {
	cfg := api.DefaultConfig()
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr != "" {
		cfg.Address = addr
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	return &ConsulController{
		client: *client,
		addr:   addr,
		config: cfg,
	}
}
