package main

import (
	"time"

	"github.com/cloudwebrtc/nats-discovery/pkg/discovery"
	"github.com/cloudwebrtc/nats-discovery/pkg/registry"
	"github.com/nats-io/nats.go"
	log "github.com/pion/ion-log"
)

func init() {
	log.Init("info")
}

//setupConnOptions default conn opts.
func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second
	connectTimeout := 5 * time.Second

	opts = append(opts, nats.Timeout(connectTimeout))
	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectHandler(func(nc *nats.Conn) {
		if !nc.IsClosed() {
			log.Infof("Disconnected, will attempt reconnects for %.0fm", totalWait.Minutes())
		}
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Infof("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		if !nc.IsClosed() {
			log.Errorf("Exiting: no servers available")
		} else {
			log.Errorf("Exiting")
		}
	}))
	return opts
}

func main() {
	natsURL := nats.DefaultURL

	opts := []nats.Option{nats.Name("nats-discovery discovery server")}
	opts = setupConnOptions(opts)
	// Connect to the NATS server.
	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		log.Errorf("%v", err)
		return
	}

	reg, err := registry.NewRegistry(nc)
	if err != nil {
		log.Errorf("%v", err)
		return
	}
	reg.Listen(func(action discovery.Action, node discovery.Node) (bool, error) {
		//Add authentication here
		log.Infof("handle Node: %v, %v", action, node)
		//return false, fmt.Errorf("reject action: %v", action)
		return true, nil
	}, func(service string, params map[string]interface{}) ([]discovery.Node, error) {
		//Add load balancing here.
		log.Infof("handle get nodes: service %v, params %v", service, params)
		return reg.GetNodes(service)
	})

	select {}
}
