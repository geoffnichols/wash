package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/puppetlabs/wash/plugin"
	log "github.com/sirupsen/logrus"
)

type containers struct {
	plugin.EntryBase
	client *client.Client
}

// LS
func (cs *containers) LS(ctx context.Context) ([]plugin.Entry, error) {
	containers, err := cs.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	log.Debugf("Listing %v containers in %v", len(containers), cs)
	keys := make([]plugin.Entry, len(containers))
	for i, inst := range containers {
		keys[i] = &container{
			EntryBase: plugin.NewEntry(inst.ID),
			client:    cs.client,
			startTime: time.Unix(inst.Created, 0),
		}
	}
	return keys, nil
}