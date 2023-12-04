package structs

import (
	"fmt"
	"reflect"
	"strings"
)

type Locations struct {
	Id              uint64   `json:"id" binding:"required"`       // locations' id
	AnyTypeLocation any      `json:"location" binding:"required"` // Indicate the coordinates that will be used for route optimization. Every coordinate is separated by “|”. The coordinate format is latitude,longitude . Example: 1.29360227,103.80828989|1.29360227,103.8273062
	Approaches      []string `json:"approaches"`
	Location        string   `json:"location_str" swaggerignore:"true"`
}


func (l *Locations) ConvertLocation() error {
	v := reflect.ValueOf(l.AnyTypeLocation)
	switch v.Kind() {
	case reflect.String:
		l.Location = v.String()
	case reflect.Slice:
		result := make([]string, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if itemType := reflect.TypeOf(item.Interface()).Kind(); itemType != reflect.String {
				return fmt.Errorf("unable to convert %s to location string", itemType)
			}
			str, ok := item.Interface().(string)
			if !ok {
				return fmt.Errorf("%s is not a valid location format", str)
			}
			result = append(result, str)
		}
		l.Location = strings.Join(result, "|")
	default:
		return fmt.Errorf("%s is not a valid location format", v)
	}
	return nil
}

type Job struct {
	Id            uint64     `json:"id" binding:"required"`             // Indicate the job id. It cannot be duplicated to other Job’s id.
	LocationIndex uint64     `json:"location_index" binding:"required"` // Indicate the index of location in Locations. The valid value range is [0, length of locations)
	Service       *uint64    `json:"service"`                           // Describe the job service duration. The unit is in second. Default value is 0.
	Delivery      []uint64   `json:"delivery"`                          // Describe multidimensional quantities for delivery. The amount of delivery will be added to the assigned vehicle’s initial load.
	Pickup        []uint64   `json:"pickup"`                            // Describe multidimensional quantities for pickup.
	TimeWindows   [][]uint64 `json:"time_windows"`                      // Describe available periods for this job.
	Skills        []uint64   `json:"skills"`                            // Describe mandatory skills for this job
	Priority      *uint64    `json:"priority"`                          // Describe the priority of this job. The valid value is in range of [0, 100]. Its default value is 0. Priority only decides whether this job will be assigned, but has nothing to do with the sequence of job.
	Setup         *uint64    `json:"setup"`                             // Describe the job setup duration. The unit is in second. Its default value is 0.
	TimeWindow    [][]uint64 `json:"time_window" swaggerignore:"true"`
	Description   *string    `json:"description"` // Describe this job
}

type Vehicle struct {
	Id          uint64       `json:"id" binding:"required"` // Describe this vehicle’s id. It cannot be duplicated to other vehicle’s id.
	StartIndex  *uint64      `json:"start_index"`           // Indicate the index of vehicle starting point in Locations. The valid value range is [0, length of locations)
	EndIndex    *uint64      `json:"end_index"`             // Indicate the index of vehicle ending point in Locations. The valid value range is [0, length of locations)
	Capacity    []int64      `json:"capacity"`              // Describe multidimensional quantities of capacity
	TimeWindow  []uint64     `json:"time_window"`           // Describe the vehicle available time period
	Skills      []uint64     `json:"skills"`                // Describe the skills driver/vehicle has
	Breaks      []Break      `json:"breaks"`                // Describe the breaks the driver will take
	MaxTasks    *uint64      `json:"max_tasks"`             // Describe the max tasks can be assigned to this vehicle
	Costs       VehicleCosts `json:"costs"`                 // Describe the cost configurations for a vehicle
	Depot       *uint64      `json:"depot"`                 // Describe the depot assigned to this vehicle
	Description *string      `json:"description"`           // Describe this vehicle
	SpeedFactor *float64     `json:"speed_factor,omitempty"`
}

type Shipment struct {
	Pickup   *ShipmentStep `json:"pickup" binding:"required"`   // Describe the pickup point for this shipment
	Delivery *ShipmentStep `json:"delivery" binding:"required"` // Describe the delivery point for this shipment
	Amount   []uint64      `json:"amount"`                      // Describe multidimensional quantities
	Skills   []uint64      `json:"skills"`                      // Describe the mandatory skills for this shipment
	Priority *uint64       `json:"priority"`                    // Describe the priority of this shipment. The valid value is in range of [0, 100]. Its default value is 0. Priority only decides whether this shipment will be assigned, but has nothing to do with the sequence of shipments.
}

type ShipmentStep struct {
	Id            uint64     `json:"id" binding:"required"`             // Indicate the id of this shipment_step. An error will be reported when there’re duplicate ids for pickup/delivery
	LocationIndex uint64     `json:"location_index" binding:"required"` // Indicate the position of this shipment_step. The valid range of value is [0, length of locations)
	Service       uint64     `json:"service"`                           // Describe the service duration of this shipment_step. Default value is 0. The unit is in seconds.
	TimeWindows   [][]uint64 `json:"time_windows"`                      // Describe the available periods for this shipment_step
	TimeWindow    [][]uint64 `json:"time_window" swaggerignore:"true"`
	Description   *string    `json:"description"` // Describe this shipment step
	Setup         *uint64    `json:"setup"`
}


type Depot struct {
	Id            uint64  `json:"id" binding:"required"`             // Describe the id of this depot
	LocationIndex uint64  `json:"location_index" binding:"required"` // Describe the location of depot
	Description   *string `json:"description"`                       // Add a description for this depot
}

type Unassigned struct {
	Id       uint64    `json:"id"`                 // Indicate the id of unassigned task
	Type     string    `json:"type,omitempty"`     // Describe the coordinate of the unassigned task
	Location []float64 `json:"location,omitempty"` // Describe the unassigned task type
}

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

type TimeWindow struct {
	Start float64 `json:"start" binding:"required"`
	End   float64 `json:"end" binding:"required"`
}

type Violation struct {
	Cause    *string  `json:"cause,omitempty"`    // Describe the cause of violation
	Duration *float64 `json:"duration,omitempty"` // Earliness (resp. lateness) if cause is "lead_time" (resp "delay")
}

type Summary struct {
	Cost        *uint64     `json:"cost,omitempty"`                            // Describe the cost of the routings. Right now, cost is equal to duration for we’re optimizing on shortest travel time.
	Routes      *uint64     `json:"routes,omitempty"`                          // Describe the number of routes in the solution
	Unassigned  uint64      `json:"unassigned"`                                // Describe the number of tasks not served
	Setup       *uint64     `json:"setup,omitempty"`                           // Describe total time of setups
	Service     *uint64     `json:"service,omitempty"`                         // Describe total time of service
	Duration    *float64    `json:"duration,omitempty"`                        // Describe total time for all routes
	WaitingTime *uint64     `json:"waiting_time,omitempty"`                    // Describe total waiting time for all routes
	Priority    *uint64     `json:"priority,omitempty"`                        // Describe sum of priorities in all assigned tasks.
	Violations  []Violation `json:"violations,omitempty" swaggerignore:"true"` // Describe all the violations for all routes
	Delivery    []uint64    `json:"delivery,omitempty"`                        // Describe total delivery for all routes
	Pickup      []uint64    `json:"pickup,omitempty"`                          // Describe total pickup for all routes
	Distance    float64     `json:"distance"`                                  // Describe total distance for all routes
}


// Optimization 
type VehicleCosts struct {
	Fixed    uint64  `json:"fixed"`
	PerHour  *uint64 `json:"per_hour,omitempty"`
	PerHours *uint64 `json:"per_hours,omitempty"`
}



type VehicleProfile string




type UsedObject struct {
	Locations  []uint64
	Pickups    []uint64
	Deliveries []uint64
	Starts     []uint64
	Ends       []uint64
	Skills     []uint64
}

type OptimizationOptions struct {
	Objective ObjectiveOptions `json:"objective"` // Describe the objectives for this optimization job
	Routing   RoutingOptions   `json:"routing,omitempty"`
}

type ObjectiveOptions struct {
	TravelCost        string `json:"travel_cost"`         // Select a type of cost (duration or distance)
	MinimiseNumDepots bool   `json:"minimise_num_depots"` // Choose whether to minimise the number of depots in route
}

type RoutingOptions struct {
	TrafficTimestamp *uint64 `json:"traffic_timestamp"`
	TruckSize        *string `json:"truck_size"`
	TruckWeight      *uint64 `json:"truck_weight"`
	Mode             *string `json:"mode"`
}



