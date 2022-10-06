package database

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	mock "github.com/ipfs/go-ipfs/core/mock"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	icore "github.com/ipfs/interface-go-ipfs-core"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/mitchellh/mapstructure"
)

func setupPlugins(path string) error {
	plugins, err := loader.NewPluginLoader(filepath.Join(path, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func createNode(ctx context.Context, repoPath string, m mocknet.Mocknet) (*core.IpfsNode, icore.CoreAPI, error) {
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, nil, err
	}

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTClientOption, // DHTOption
		Repo:    repo,
		Host:    mock.MockHostOption(m),
		ExtraOpts: map[string]bool{
			"pubsub": true,
		},
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, nil, err
	}

	coreAPI, err := coreapi.NewCoreAPI(node)
	if err != nil {
		return nil, nil, err
	}

	return node, coreAPI, nil
}

// func getUnixfsNode(path string) (files.Node, error) {
// 	st, err := os.Stat(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	f, err := files.NewSerialFile(path, false, st)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return f, nil
// }

func StructToMap(v interface{}) (map[string]interface{}, error) {
	vMap := &map[string]interface{}{}

	err := mapstructure.Decode(v, &vMap)
	if err != nil {
		return nil, err
	}

	return *vMap, nil
}
