package types

import (
	pb "CarpoolSharing/shared/proto/trip"
)

// Route Model
type OsrmApiResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

func (o *OsrmApiResponse) ToProto() *pb.Route {
	route := o.Routes[0]
	geometry := route.Geometry.Coordinates
	coordinates := make([]*pb.Coordinate, len(geometry))
	for i, coord := range geometry {
		coordinates[i] = &pb.Coordinate{
			Latitude:  coord[0],
			Longitude: coord[1],
		}
	}

	return &pb.Route{
		Geometry: []*pb.Geometry{
			{
				Coordinates: coordinates,
			},
		},
		Distance: route.Distance,
		Duration: route.Duration,
	}
}

// RideFare Calculation Model
type PricingConfig struct {
	PricePerUnitOfDistance float64
	PricePerMinute         float64
}

func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		PricePerUnitOfDistance: 1.5,
		PricePerMinute:         0.25,
	}
}
