package builders

import (
	"fmt"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

var nodeTypes = []string{"airport", "railway_station", "bus_station", "port"}
var nodeLocations = []string{"Центр", "Окраина", "Пригород", "Терминал 1"}

type TransportNodeBuilder struct {
	node *domain.TransportNode
}

func NewTransportNodeBuilder() *TransportNodeBuilder {
	id := domain.NodeID(uuid.New())
	name := fmt.Sprintf("Транспортный узел_%d", rnd.Intn(10000))
	nodeType := nodeTypes[rnd.Intn(len(nodeTypes))]
	coords := domain.NewCoordinates(40.0+rnd.Float64()*15.0, -10.0+rnd.Float64()*40.0)
	location := nodeLocations[rnd.Intn(len(nodeLocations))]

	cityID := domain.CityID(uuid.New())
	imageID := DefaultImageID

	node, err := domain.NewTransportNode(
		id, name, nodeType, coords, location, cityID, imageID,
	)
	if err != nil {
		panic(fmt.Errorf("builders: failed to create mock transport node: %w", err))
	}

	return &TransportNodeBuilder{node: node}
}

func (b *TransportNodeBuilder) WithID(id domain.NodeID) *TransportNodeBuilder {
	b.node.SetId(id)
	return b
}

func (b *TransportNodeBuilder) WithName(name string) *TransportNodeBuilder {
	b.node.SetName(name)
	return b
}

func (b *TransportNodeBuilder) WithType(nodeType string) *TransportNodeBuilder {
	b.node.SetNodeType(nodeType)
	return b
}

func (b *TransportNodeBuilder) WithCityID(cityID domain.CityID) *TransportNodeBuilder {
	b.node.SetCityID(cityID)
	return b
}

func (b *TransportNodeBuilder) Build() *domain.TransportNode {
	return b.node
}
