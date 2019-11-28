package consul

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
)

type Client struct {
	consul *consul.Client
}

// NewConsul returns a Client interface for given consul address
func NewConsulClient(addr string) (*Client, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	c, err := consul.NewClient(config)
	if err != nil {
		return &Client{}, err
	}
	return &Client{consul: c}, nil
}

// Register a service with consul local agent - note the tags to define path-prefix is to be used.
func (c *Client) Register(id, name, host string, port int, health string) error {
	reg := &consul.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Port:    port,
		Address: host,
		Check: &consul.AgentServiceCheck{
			CheckID:       id,
			Name:          "HTTP API health",
			HTTP:          health,
			TLSSkipVerify: true,
			Method:        "GET",
			Interval:      "10s",
			Timeout:       "1s",
		},
		Tags: []string{
			"traefik.enable=true",
			"traefik.frontend.rule=Host:localhost;PathPrefixStrip:/" + name,
			"traefik.frontend.entryPoints=http",
			"traefik.backend=" + name,
		},
	}
	return c.consul.Agent().ServiceRegister(reg)
}

// DeRegister a service with consul local agent
func (c *Client) DeRegister(id string) error {
	return c.consul.Agent().ServiceDeregister(id)
}

// Service return a service
func (c *Client) Service(serviceName, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := c.consul.Health().Service(serviceName, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", serviceName)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}

func (c *Client) ServiceAddress(serviceName string) (string, error) {
	srvc, _, err := c.Service(serviceName, "")
	if err != nil {
		return "", err
	}

	address := srvc[0].Service.Address
	port := srvc[0].Service.Port
	return fmt.Sprintf("http://%s:%v", address, port), nil
}
