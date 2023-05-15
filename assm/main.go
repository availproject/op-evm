package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/secrets/awsssm"
	"github.com/0xPolygon/polygon-edge/secrets/helper"
	edgetypes "github.com/0xPolygon/polygon-edge/types"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/go-hclog"
)

//go:embed genesis-defaults.json
var genesisDefaultsFile []byte

// ETH token value
var ETH = big.NewInt(1000000000000000000)

const (
	defaultBalance   = 1_000_000_000
	genesisJsonKey   = "genesis.json"
	nodesInfoJsonKey = "nodesInfo.json"

	EnvAwsRegion    = "AWS_REGION"
	EnvSsmParamPath = "SSM_PARAM_PATH"
	EnvNamespace    = "SSM_NAMESPACE"
	EnvS3BucketName = "S3_BUCKET_NAME"
	EnvTotalNodes   = "TOTAL_NODES"
	EnvBalance      = "BALANCE"
)

func main() {
	h := &handler{}
	err := json.Unmarshal(genesisDefaultsFile, &h.chainDefaults)
	if err != nil {
		log.Fatal(err)
	}
	region, ok := os.LookupEnv(EnvAwsRegion)
	if !ok {
		log.Fatal("AWS_REGION env var required but wasn't provided")
	}
	ssmParamPath, ok := os.LookupEnv(EnvSsmParamPath)
	if !ok {
		log.Fatal("SSM_PARAM_PATH env var required but wasn't provided")
	}
	namespace, ok := os.LookupEnv(EnvNamespace)
	if !ok {
		log.Fatal("SSM_NAMESPACE env var required but wasn't provided")
	}
	h.ssmFactory = newSecretsManagerFactory(region, ssmParamPath, namespace)
	s3BucketName, ok := os.LookupEnv(EnvS3BucketName)
	if !ok {
		log.Fatal("S3_BUCKET_NAME env var required but wasn't provided")
	}
	if h.storage, err = NewStorage(region, s3BucketName); err != nil {
		log.Fatal(err)
	}
	totalNodesRaw, ok := os.LookupEnv(EnvTotalNodes)
	if !ok {
		log.Fatal("TOTAL_NODES env var required but wasn't provided")
	}
	if h.totalNodes, err = strconv.Atoi(totalNodesRaw); err != nil {
		log.Fatal("error parsing TOTAL_NODES as int", err)
	}
	balanceRaw, ok := os.LookupEnv(EnvBalance)
	if !ok {
		h.balance = defaultBalance
	} else if h.balance, err = strconv.ParseInt(balanceRaw, 10, 0); err != nil {
		log.Fatal("error parsing BALANCE as int", err)
	}

	lambda.Start(h.handle)
}

type NodeInfo struct {
	NodeIP   string `json:"node_ip"`
	NodeDNS  string `json:"node_dns"`
	NodeName string `json:"node_name"`
	NodePort string `json:"node_port"`
	NodeType string `json:"node_type"`
}

type Response struct {
	Message string `json:"message"`
}

type handler struct {
	totalNodes    int
	balance       int64
	chainDefaults chain.Chain
	storage       *Storage
	ssmFactory    secretsManagerFactory
}

func (h *handler) handle(ctx context.Context, nodeReq *NodeInfo) (*Response, error) {
	nodeReq.NodeIP = strings.Trim(nodeReq.NodeIP, " ")
	nodeReq.NodeDNS = strings.Trim(nodeReq.NodeDNS, " ")
	nodeReq.NodeName = strings.Trim(nodeReq.NodeName, " ")
	nodeReq.NodePort = strings.Trim(nodeReq.NodePort, " ")
	nodeReq.NodeType = strings.Trim(nodeReq.NodeType, " ")

	if nodeReq.NodeIP == "" && nodeReq.NodeDNS == "" {
		return nil, fmt.Errorf("node_ip or node_dns is required but wasn't provided")
	}
	if nodeReq.NodeName == "" {
		return nil, fmt.Errorf("node_name is required but wasn't provided")
	}
	if nodeReq.NodePort == "" {
		return nil, fmt.Errorf("node_port is required but wasn't provided")
	}
	if nodeReq.NodeType == "" {
		return nil, fmt.Errorf("node_type is required but wasn't provided")
	}

	// Fetch existing nodes data.
	var nodes []*NodeInfo
	if err := h.storage.Get(ctx, nodesInfoJsonKey, &nodes); err != nil {
		return nil, fmt.Errorf("could not fetch data from S3 bucket err=%w", err)
	}

	nodes = append(nodes, nodeReq)

	// Save the nodes to s3 until info about all nodes was received.
	if h.totalNodes > len(nodes) {
		if err := h.storage.Set(ctx, nodesInfoJsonKey, nodes); err != nil {
			return nil, fmt.Errorf("could not write data to S3 err=%w", err)
		}
		return &Response{
			Message: fmt.Sprintf("Node with number %d and name %s info successfully saved. Need a total of %d nodes.",
				len(nodes), nodeReq.NodeName, h.totalNodes),
		}, nil
	}
	for _, node := range nodes {
		ssm, err := h.ssmFactory(node.NodeName)
		if err != nil {
			return nil, fmt.Errorf("unable to init secrets manager err=%w", err)
		}

		// Once we got information about all nodes proceed to generate the genesis file
		nodeID, err := helper.LoadNodeID(ssm)
		if err != nil {
			return nil, fmt.Errorf("unable to load node id from ssm err=%w", err)
		}
		if nodeID == "" {
			return nil, fmt.Errorf("node id not found in secrets")
		}
		validatorAddress, err := helper.LoadValidatorAddress(ssm)
		if err != nil {
			return nil, fmt.Errorf("unable to load validator address from ssm err=%w", err)
		}
		if validatorAddress == edgetypes.ZeroAddress {
			return nil, fmt.Errorf("validator address not found in secrets")
		}
		h.chainDefaults.Genesis.Alloc[validatorAddress] = &chain.GenesisAccount{
			Balance: big.NewInt(0).Mul(big.NewInt(h.balance), ETH),
		}
		if node.NodeType == "bootstrap-sequencer" || node.NodeType == "sequencer" {
			if node.NodeIP != "" {
				h.chainDefaults.Bootnodes = append(h.chainDefaults.Bootnodes,
					fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", node.NodeIP, node.NodePort, nodeID))
			} else if node.NodeDNS != "" {
				h.chainDefaults.Bootnodes = append(h.chainDefaults.Bootnodes,
					fmt.Sprintf("/dns/%s/tcp/%s/p2p/%s", node.NodeDNS, node.NodePort, nodeID))
			}
		}
		if err := h.storage.Set(ctx, genesisJsonKey, h.chainDefaults); err != nil {
			return nil, fmt.Errorf("unable to write genesis file to the bucket err=%w", err)
		}
		// Write the last node data to aws regardless for debugging and consistancy
		if err := h.storage.Set(ctx, nodesInfoJsonKey, nodes); err != nil {
			return nil, fmt.Errorf("could not write data to S3 err=%w", err)
		}
	}

	return &Response{Message: "Genesis file successfully created and uploaded to S3"}, nil
}

type secretsManagerFactory func(name string) (secrets.SecretsManager, error)

func newSecretsManagerFactory(region, ssmParamPath, namespace string) func(name string) (secrets.SecretsManager, error) {
	return func(name string) (secrets.SecretsManager, error) {
		return awsssm.SecretsManagerFactory(&secrets.SecretsManagerConfig{
			Name:      name,
			Namespace: namespace,
			Type:      secrets.AWSSSM,
			Extra: map[string]interface{}{
				"region":             region,
				"ssm-parameter-path": ssmParamPath,
			},
		}, &secrets.SecretsManagerParams{Logger: hclog.NewNullLogger()})
	}
}

type Storage struct {
	s3         *s3.Client
	bucketName string
}

func NewStorage(region, bucketName string) (*Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("could not load new aws config err=%w", err)
	}

	return &Storage{
		s3:         s3.NewFromConfig(cfg),
		bucketName: bucketName,
	}, nil
}

func (a *Storage) Set(ctx context.Context, key string, data interface{}) error {
	buff := &bytes.Buffer{}
	if err := json.NewEncoder(buff).Encode(data); err != nil {
		return fmt.Errorf("unable to encode json err=%w", err)
	}

	_, err := a.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &a.bucketName,
		Key:    &key,
		Body:   buff,
	})
	if err != nil {
		return fmt.Errorf("could not put object to S3 err=%w", err)
	}

	return nil
}

func (a *Storage) Get(ctx context.Context, key string, v interface{}) error {
	s3Object, err := a.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &a.bucketName,
		Key:    &key,
	})
	if err != nil {
		var noSuchKey *awstypes.NoSuchKey
		if errors.As(err, &noSuchKey) {
			fmt.Println("The file was not uploaded yet, key=", key)
			return nil
		}

		return fmt.Errorf("could not fetch S3 object bucket=%v key=%s err=%w", a.bucketName, key, err)
	}

	return json.NewDecoder(s3Object.Body).Decode(v)
}
