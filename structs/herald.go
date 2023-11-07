package structs

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"herald/config"
	"herald/storage/remote"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Optimization 
type PubsubSchema struct {
	APIKey    string         `json:"api_key"`
	Endpoint  string         `json:"endpoint"`
	TimeStamp uint64         `json:"timestamp"`
	Metadata  PubsubMetadata `json:"metadata"`
	OrgId     string         `json:"org_id"`
	Namespace string         `json:"k8s_namespace"`
	Cluster   string         `json:"cluster"`
	Sku       string         `json:"sku"`
}


// Optimization
type PubsubMetadata struct {
	LocationHashes []string `json:"location_hashes"`
}


// Optimization
type OptimizationStore struct {
	StartTime   int64
	Description *string
	VRoomResult []byte
	Error       string
	Key         string
	IsEndState  bool
}

// swagger:ignore


type VRoomStep struct { // will change to herald
	Type          *string     `json:"type,omitempty"`                            // Describe the task type of this step
	Arrival       *float64    `json:"arrival,omitempty"`                         // Describe the arrival time at this step
	Duration      *float64    `json:"duration,omitempty"`                        // Describe the duration to this step. (The duration is accumulated here which means it includes the time spent on previous steps)
	Setup         *uint64     `json:"setup,omitempty"`                           // Describe the setup duration at this step
	Service       *uint64     `json:"service,omitempty"`                         // Describe the service duration at this step
	WaitingTime   *uint64     `json:"waiting_time,omitempty"`                    // Describe the waiting time at this step
	Violations    []Violation `json:"violations,omitempty" swaggerignore:"true"` // Describe the violations in this step
	Location      []float64   `json:"location,omitempty"`                        // Describe the coordinate at this step
	Id            *uint64     `json:"id,omitempty"`                              // Describe the id of this task
	Load          []float64   `json:"load,omitempty"`                            // Describe the load of the vehicle after the completion of this step
	Description   *string     `json:"description,omitempty"`                     // Describe this step
	LocationIndex *uint64     `json:"location_index" binding:"required"`
	Distance      *uint64     `json:"distance,omitempty"`
}

type VRoomRoute struct { // will change to herald
	Vehicle     *uint64     `json:"vehicle"`                // Describe the id of assigned vehicle
	Cost        uint64      `json:"cost"`                   // Describe the cost of this route. Right now it is equal to duration
	Steps       []VRoomStep `json:"steps"`                  // Describe the steps in this route
	Setup       *uint64     `json:"setup,omitempty"`        // Describe the total setup time for this route
	Service     *uint64     `json:"service,omitempty"`      // Describe the total service time for this route
	Duration    *uint64     `json:"duration"`               // Describe the duration of this route
	WaitingTime *uint64     `json:"waiting_time,omitempty"` // Describe the total waiting time for this route
	Priority    *uint64     `json:"priority,omitempty"`     // Describe the sum of priorities for this route
	Violations  []Violation `json:"violations,omitempty" swaggerignore:"true"`
	Delivery    []uint64    `json:"delivery,omitempty"`    // Describe the total deliveries in this route
	Pickup      []uint64    `json:"pickup,omitempty"`      // Describe the total pickups in this route
	Distance    *float64    `json:"distance"`              // Describe the total distance in this route
	Geometry    *string     `json:"geometry,omitempty"`    // The polyline for this route
	Description *string     `json:"description,omitempty"` // The description for the assigned vehicle
}

type VRoomResult struct { // will change to herald
	Code       *uint8       `json:"code,omitempty"`       // 0: no error, 1: internal error, 2: input error, 3: routing error
	Error      *string      `json:"error,omitempty"`      // Describe the error when there is
	Summary    *Summary     `json:"summary"`              // Summarize the solution
	Unassigned []Unassigned `json:"unassigned,omitempty"` // Describe the unassigned tasks
	Routes     []VRoomRoute `json:"routes"`               // Describe the optimization routes
}

type OptimizationPostInput struct {
	Locations   Locations           `json:"locations" binding:"required"` // Describes the locations which will be used in optimization
	Jobs        []Job               `json:"jobs"`                         // Describes the jobs to be assigned to vehicles
	Vehicles    []Vehicle           `json:"vehicles" binding:"required"`  // Describes the vehicles
	Shipments   []Shipment          `json:"shipments"`                    // Describes shipments to be assigned to vehicles
	Description *string             `json:"description"`                  // Describes this optimization task
	Options     OptimizationOptions `json:"options"`                      // Describes the optimization options
	Depots      []Depot             `json:"depots"`                       // Describes the locations of depots
	Depot       []Depot             `json:"depot" swaggerignore:"true"`
	Mode        *string             `json:"mode"`
	CostMatrix  [][]uint64          `json:"cost_matrix" swaggerignore:"true"`
}

type GatewayHeader struct {
	Referer            string `header:"referer" json:"referer"`
	NbGatewayTrackInfo string `header:"nb-gateway-track-info" json:"nb-gateway-track-info"`
}


type TrackInfo struct {
	EndpointName            string             `json:"endpoint_name"`
	SinkTo                  *string            `json:"sink_to"`
	SaasLabels              *map[string]string `json:"saas_labels"`
	UserAgent               *string            `json:"user_agent"`
	Source                  *string            `json:"source"`
	IsLocal                 bool               `json:"is_local"`
	IsInteral               bool               `json:"is_interal"`
	EndpointType            *interface{}       `json:"endpoint_type"`
	UnderlyingElementsCount *uint64            `json:"underlying_elements_count"`
	ReadyToForward          bool               `json:"ready_to_forward"`
	ForwardScheme           *string            `json:"forward_scheme"`
	ProxyElapseSeconds      *float32           `json:"proxy_elapse_seconds"`
	Method                  string             `json:"method"`
	ElapseSeconds           float32            `json:"elapse_seconds"`
	RequestId               string             `json:"request_id"`
}

func (input *OptimizationPostInput) GenJobID(apikey string, jobIDPrefix string) (string, error) {
	inputByte, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	h := md5.New()
	io.WriteString(h, string(inputByte))
	io.WriteString(h, apikey)
	hash := h.Sum(nil)
	id := hex.EncodeToString(hash[:])
	isErrorJob := ifErrorJob(id)
	if isErrorJob || !config.Conf.CacheId {
		// allow user to recreate error job instead of returning the same ID
		io.WriteString(h, strconv.FormatInt(time.Now().UnixMilli(), 10))
		hash = h.Sum(nil)
		id = hex.EncodeToString(hash[:])
	}
	return jobIDPrefix + id, nil
}

func ifErrorJob(id string) bool {
	result, err := remote.Client.Get("Optimization_" + id)
	if err != nil {
		return false
	}
	var content OptimizationStore
	err = json.Unmarshal([]byte(result), &content)
	if err != nil {
		return false
	}
	if len(content.Error) > 0 {
		return true
	}
	return false
}

type OptimizationPostQuery struct {
	Key string `form:"key" binding:"required"`
}

type OptimizationPostOutput struct {
	Id      string   `json:"id" binding:"required"`      // Describe the id which will be used in optimization GET to get the result of optimization
	Message string   `json:"message" binding:"required"` // Describe the request
	Status  string   `json:"status" binding:"required"`  // Describe the request status
	Warning []string `json:"warning,omitempty"`          // Display the potential lints in input fields
}

type OptimizationGetInput struct {
	Key string `form:"key"`
	Id  string `form:"id" binding:"required"`
}

type OptimizationGetOutput struct {
	Description string      `json:"description,omitempty"`      // It will be returned when it is given in optimization POST locations’ description.
	Result      VRoomResult `json:"result" binding:"required"`  // Describe the optimization routing result
	Status      string      `json:"status" binding:"required"`  // Describe the error happens during processing data
	Message     string      `json:"message" binding:"required"` // Describe process status
}



type SimpleErrorResp struct {
	Message  string   `json:"message"`
	Warnings []string `json:"warnings,omitempty"`
}


type VehicleRoutingMsg struct {
	Jobs      []VRoomJob          `json:"jobs"`
	Shipments []VRoomShipment     `json:"shipments,omitempty"`
	Vehicles  []VRoomVehicle      `json:"vehicles" binding:"required"`
	Matrices  map[string]Matrix   `json:"matrices" binding:"required"`
	Depots    []Depot             `json:"depots,omitempty"`
	Options   OptimizationOptions `json:"options"`
}
type Matrix struct {
	Durations [][]uint64 `json:"durations,omitempty"`
	Costs     [][]uint64 `json:"costs,omitempty"`
}

type VRoomShipment struct { // will change to Herald
	Pickup   *VRoomShipmentStep `json:"pickup,omitempty"`
	Delivery *VRoomShipmentStep `json:"delivery,omitempty"`
	Amount   []uint64           `json:"amount,omitempty"`
	Skills   []uint64           `json:"skills,omitempty"`
	Priority *uint64            `json:"priority,omitempty"`
}

type VRoomShipmentStep struct { // Will change to Herald
	Id            uint64     `json:"id" binding:"required"`
	Description   string     `json:"description,omitempty"`
	Location      []float64  `json:"location"`
	LocationIndex uint64     `json:"location_index"`
	Setup         *uint64    `json:"setup,omitempty"`
	Service       *uint64    `json:"service,omitempty"`
	TimeWindows   [][]uint64 `json:"time_windows,omitempty"`
}

type VRoomJob struct {
	Id            uint64     `json:"id" binding:"required"`
	Description   string     `json:"description,omitempty"`
	Location      []float64  `json:"location" binding:"required"`
	LocationIndex uint64     `json:"location_index" binding:"required"`
	Setup         *uint64    `json:"setup,omitempty"`
	Service       *uint64    `json:"service,omitempty"`
	Delivery      []uint64   `json:"delivery,omitempty"`
	Pickup        []uint64   `json:"pickup,omitempty"`
	Skills        []uint64   `json:"skills,omitempty"`
	Priority      *uint64    `json:"priority,omitempty"`
	TimeWindows   [][]uint64 `json:"time_windows,omitempty"`
}

type VRoomVehicle struct { // will change to herald
	Id          uint64         `json:"id" binding:"required"`
	Profile     VehicleProfile `json:"profile" binding:"required"`
	Description string         `json:"description,omitempty"`
	Start       []float64      `json:"start,omitempty"`
	StartIndex  *uint64        `json:"start_index,omitempty"`
	End         []float64      `json:"end,omitempty"`
	EndIndex    *uint64        `json:"end_index,omitempty"`
	Capacity    []int64        `json:"capacity,omitempty"`
	Skills      []uint64       `json:"skills,omitempty"`
	TimeWindow  []uint64       `json:"time_window,omitempty"`
	Breaks      []Break        `json:"breaks,omitempty"`
	SpeedFactor *float64       `json:"speed_factor,omitempty"`
	MaxTasks    *uint64        `json:"max_tasks,omitempty"`
	Steps       []VehicleStep  `json:"steps,omitempty"`
	Costs       VehicleCosts   `json:"costs"`
	Depot       *uint64        `json:"depot,omitempty"`
}

type Break struct {
	Id          uint64     `json:"id" binding:"required"`           // Indicate the id for this break. It cannot be duplicated to other break’s id for the same vehicle.
	TimeWindows [][]uint64 `json:"time_windows" binding:"required"` // Describe the possible periods to take break
	Service     uint64     `json:"service"`                         // Describe the break duration. Its default value is 0. The unit is in second.
	Description string     `json:"description,omitempty"`           // Describe this break
}

type VehicleStep struct {
	Type          string `json:"type" binding:"required"`
	Id            uint64 `json:"id,omitempty"`
	ServiceAt     uint64 `json:"service_at,omitempty"`
	ServiceAfter  uint64 `json:"service_after,omitempty"`
	ServiceBefore uint64 `json:"service_before,omitempty"`
}


