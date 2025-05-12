package main

import (
	"fmt"
	"math/rand"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerNetwork "github.com/docker/docker/api/types/network"

	"deniable-im/im-sim/internal/types"
	"deniable-im/im-sim/pkg/client"
	"deniable-im/im-sim/pkg/container"
	"deniable-im/im-sim/pkg/image"
	"deniable-im/im-sim/pkg/network"
	Behavior "deniable-im/im-sim/pkg/simulation/behavior"
	Simulator "deniable-im/im-sim/pkg/simulation/simulator"
	User "deniable-im/im-sim/pkg/simulation/simulator/user"
	Types "deniable-im/im-sim/pkg/simulation/types"
)

func main() {
	dockerClient, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	defer dockerClient.Close()

	// Build redis image
	_, err = image.NewImage(dockerClient, "./cmd/denim-sim/", &image.Options{
		PullOpt: &image.PullOptions{
			RefStr: "redis:latest",
		},
	})
	if err != nil {
		panic(err)
	}

	// Build DB image
	_, err = image.NewImage(dockerClient, "./cmd/denim-sim/", &image.Options{
		BuildOpt: &dockerTypes.ImageBuildOptions{
			Dockerfile: "Dockerfile.postgres",
			Tags:       []string{"denim-postgres"},
		},
	})
	if err != nil {
		panic(err)
	}

	// Build server image
	_, err = image.NewImage(dockerClient, "./cmd/denim-sim/", &image.Options{
		BuildOpt: &dockerTypes.ImageBuildOptions{
			Dockerfile: "Dockerfile.server",
			Tags:       []string{"denim-server"},
		},
	})
	if err != nil {
		panic(err)
	}

	// Build client image
	_, err = image.NewImage(dockerClient, "./cmd/denim-sim/", &image.Options{
		BuildOpt: &dockerTypes.ImageBuildOptions{
			Dockerfile: "Dockerfile.client",
			Tags:       []string{"denim-client"},
		},
	})
	if err != nil {
		panic(err)
	}

	// Create network for DB, cache and server
	networkBackend := network.NewNetwork(dockerClient, "backend", network.Options{
		Driver: "bridge",
	})

	// Create network that supports 2046 IPs
	networkOptions := network.Options{
		Driver: "macvlan",
		IPAM: &dockerNetwork.IPAM{
			Config: []dockerNetwork.IPAMConfig{
				{
					Subnet:  "10.10.240.0/20",
					IPRange: "10.10.248.0/21",
					Gateway: "10.10.248.1",
				},
			},
		},
	}

	// Create network
	networkIMvlan := network.NewNetwork(dockerClient, "IMvlan", networkOptions)

	// Setup redis
	cache, err := container.NewContainer(
		dockerClient,
		"redis:latest",
		"denim-redis",
		&container.Options{
			Connections: network.NewConnections(networkBackend),
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
			NetworkConfig: &dockerNetwork.NetworkingConfig{
				EndpointsConfig: map[string]*dockerNetwork.EndpointSettings{
					"backend": {
						Aliases: []string{"redis"},
					},
				},
			},
		})
	if err != nil {
		panic(err)
	}

	if err := cache.Start(); err != nil {
		panic(err)
	}

	// Setup DB
	db, err := container.NewContainer(
		dockerClient,
		"denim-postgres",
		"denim-postgres",
		&container.Options{
			Connections: network.NewConnections(networkBackend),
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
			NetworkConfig: &dockerNetwork.NetworkingConfig{
				EndpointsConfig: map[string]*dockerNetwork.EndpointSettings{
					"backend": {
						Aliases: []string{"db"},
					},
				},
			},
		})
	if err != nil {
		panic(err)
	}

	if err := db.Start(); err != nil {
		panic(err)
	}

	// Setup server
	serverIP := "10.10.248.2"
	server, err := container.NewContainer(
		dockerClient,
		"denim-server",
		"denim-server",
		&container.Options{
			Connections: network.NewConnections(
				networkBackend,
				network.NewAddrMapping(networkIMvlan, serverIP),
			),
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
			NetworkConfig: &dockerNetwork.NetworkingConfig{
				EndpointsConfig: map[string]*dockerNetwork.EndpointSettings{
					"IMvlan": {
						Aliases: []string{"server"},
					},
				},
			},
		})
	if err != nil {
		panic(err)
	}

	if err := server.Start(); err != nil {
		panic(err)
	}

	// Setup clients
	var images []types.Pair[string, string]
	for i := range [100]int{} {
		images = append(images,
			types.Pair[string, string]{
				Fst: "denim-client",
				Snd: fmt.Sprintf("denim-client-%d", i),
			})
	}

	clientContainers, err := container.NewContainerSlice(
		dockerClient,
		images,
		&container.Options{
			Connections: network.NewConnections(networkIMvlan),
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
		})
	if err != nil {
		panic(err)
	}

	reservedIP := []string{"10.10.248.2"}
	clientContainers, err = container.AssignIP(clientContainers, reservedIP, *networkIMvlan)
	if err != nil {
		panic(err)
	}

	container.StartContainers(clientContainers)

	time.Sleep(5 * time.Second) //IMPORTANT SLEEP FOR ARTHUR'S MACHINE

	networkName := fmt.Sprintf("dm-%v", networkIMvlan.ID[:12])

	nextfunc := func(sht *Behavior.SimpleHumanTraits) float64 {
		var next float64 = 10
		if sht.IsBursting() {
			next = next * sht.BurstModifier
			sht.DeniableBurstSize -= 1
		}

		return float64(sht.GetRandomizer().Int31n(int32(next)))
	}

	r := rand.New(rand.NewSource(42069))

	user_count := 100
	users := make([]*User.SimulatedUser, user_count)
	// f := fuzz.NewWithSeed(6942069).NilChance(0)
	traits := Behavior.GenerateRealisticSimpleHumanTraits(user_count, nil, nextfunc)
	for i := 0; i < user_count; i++ {
		traits[i].ResponseProb += 0.2
		// traits[i].DeniableBurstSize = 0
		users[i] = &User.SimulatedUser{Behavior: traits[i], User: &Types.SimUser{ID: int32(i), Nickname: fmt.Sprintf("%v", i)}, Client: clientContainers[i]}
	}

	User.CreateCommunicationNetwork(users, 2, 4, r)
	User.CreateDeniableNetwork(users, 1, 2, r)

	println("Starting simulation")
	Simulator.SimulateTraffic(users, 2*3600, networkName)
}
