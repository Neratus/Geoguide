package domain

import (
	"strings"
)

const (
	TransportNodeTypeAirport      = "airport"
	TransportNodeTypeTrainStation = "train_station"
	TransportNodeTypeBusStation   = "bus_station"
	TransportNodeTypePort         = "port"
)

type TransportNode struct {
	id          NodeID
	name        string
	nodeType    string
	coordinates Coordinates
	address     string
	cityID      CityID
	imageID     ImageID
}

func (t *TransportNode) GetId() NodeID {
	return t.id
}

func (t *TransportNode) GetName() string {
	return t.name
}

func (t *TransportNode) GetNodeType() string {
	return t.nodeType
}

func (t *TransportNode) GetCoordinates() Coordinates {
	return t.coordinates
}

func (t *TransportNode) GetAddress() string {
	return t.address
}

func (t *TransportNode) GetCityID() CityID {
	return t.cityID
}

func (t *TransportNode) GetImageID() ImageID {
	return t.imageID
}

func (t *TransportNode) SetId(id NodeID) {
	t.id = id
}

func (t *TransportNode) SetName(name string) {
	t.name = name
}

func (t *TransportNode) SetNodeType(nodeType string) {
	t.nodeType = nodeType
}

func (t *TransportNode) SetCoordinates(coordinates Coordinates) {
	t.coordinates = coordinates
}

func (t *TransportNode) SetAddress(address string) {
	t.address = address
}

func (t *TransportNode) SetCityID(cityID CityID) {
	t.cityID = cityID
}

func (t *TransportNode) SetImageID(imageID ImageID) {
	t.imageID = imageID
}

func NewTransportNode(
	id NodeID,
	name, nodeType string,
	coordinates Coordinates,
	address string,
	cityID CityID,
	imageID ImageID,
) (*TransportNode, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyTransportNodeName
	}
	if len(name) > GetConfig().MaxTransportNodeNameLength {
		return nil, ErrTransportNodeNameTooLong
	}

	nodeType = strings.TrimSpace(nodeType)
	if nodeType == "" {
		return nil, ErrEmptyTransportNodeType
	}
	if len(nodeType) > GetConfig().MaxTransportNodeTypeLength {
		return nil, ErrTransportNodeTypeTooLong
	}
	validTypes := []string{
		TransportNodeTypeAirport,
		TransportNodeTypeTrainStation,
		TransportNodeTypeBusStation,
		TransportNodeTypePort,
	}
	valid := false
	for _, vt := range validTypes {
		if nodeType == vt {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidTransportNodeType
	}

	if len(address) > GetConfig().MaxTransportNodeAddressLength {
		return nil, ErrTransportNodeAddressTooLong
	}

	if cityID == (CityID{}) {
		return nil, ErrInvalidCityID
	}

	if err := validateCoordinates(coordinates); err != nil {
		return nil, err
	}

	return &TransportNode{
		id:          (NodeID{}),
		name:        name,
		nodeType:    nodeType,
		coordinates: coordinates,
		address:     address,
		cityID:      cityID,
		imageID:     imageID,
	}, nil
}

func NewTransportNodeFromDB(
	id NodeID,
	name, nodeType string,
	coordinates Coordinates,
	address string,
	cityID CityID,
	imageID ImageID,
) *TransportNode {
	return &TransportNode{
		id:          id,
		name:        name,
		nodeType:    nodeType,
		coordinates: coordinates,
		address:     address,
		cityID:      cityID,
		imageID:     imageID,
	}
}
