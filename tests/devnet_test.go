package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/maticnetwork/avail-settlement/consensus/avail"
)

func NewDevnetContext(jsonBlob string) (ctx *devnetContext, err error) {
	ctx = &devnetContext{}
	var instances []*Instance
	if err := json.Unmarshal([]byte(jsonBlob), &instances); err != nil {
		return nil, err
	}

	log.Printf("instances %+v\n", instances)
	ctx.urlByMechanism = map[avail.MechanismType]*url.URL{}
	for _, inst := range instances {
		nodeType := avail.MechanismType(inst.TagsAll["NodeType"])
		switch nodeType {
		case avail.Sequencer, avail.BootstrapSequencer, avail.WatchTower:
		default:
			continue
		}

		jsonRPCPort := inst.TagsAll["JsonRPCPort"]
		ctx.urlByMechanism[nodeType], err = url.Parse(fmt.Sprintf("http://%s:%s/", inst.PublicIP, jsonRPCPort))
		if err != nil {
			return nil, fmt.Errorf("unable to parse url, reason: %w", err)
		}
	}
	return ctx, nil
}

type devnetContext struct {
	urlByMechanism map[avail.MechanismType]*url.URL
}

func (d devnetContext) StopAll() {}

func (d devnetContext) GethClient() (*ethclient.Client, error) {
	u, err := d.FirstRPCURLForNodeType(avail.BootstrapSequencer)
	if err != nil {
		return nil, err
	}
	return ethclient.Dial(u.String())
}

func (d devnetContext) FirstRPCURLForNodeType(nodeType avail.MechanismType) (*url.URL, error) {
	return d.urlByMechanism[nodeType], nil
}

type Instance struct {
	PublicDns string            `json:"public_dns"`
	PublicIP  string            `json:"public_ip"`
	TagsAll   map[string]string `json:"tags_all"`
}
