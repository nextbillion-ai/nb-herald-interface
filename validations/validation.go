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


func validateTimeWindows(timeWindows [][]uint64) (bool, error) {
	var exceed24h = false
	var dayTime uint64
	dayTime = 24 * 60 * 60 // time in seconds

	for i, timeWindow := range timeWindows {
		if len(timeWindow) != 2 {
			return exceed24h, fmt.Errorf("invalid number of timestamp(s) for the time window. Each time window should contain 2 timestamps in the format [start_timestamp, end_timestamp]. Please ensure that all time windows are specified correctly")
		}
		if timeWindow[0] >= timeWindow[1] {
			return exceed24h, fmt.Errorf("invalid time window. Each time window should be in the format [start_timestamp, end_timestamp] where start_timestamp should be less/earlier than end_timestamp. Please ensure that all time windows have valid and chronological timestamps")
		}
		if timeWindow[0] > uint64(4294967295) {
			return exceed24h, fmt.Errorf("invalid timestamp value %d. Please provide a timestamp value less than 4294967295", timeWindow[0])
		}
		if timeWindow[1] > uint64(4294967295) {
			return exceed24h, fmt.Errorf("invalid timestamp value %d. Please provide a timestamp value less than 4294967295", timeWindow[1])
		}
		if timeWindow[1]-timeWindow[0] > dayTime {
			exceed24h = true
		}
		if i != 0 && timeWindow[0] <= timeWindows[i-1][1] {
			return exceed24h, fmt.Errorf("overlapping time window or unsorted time windows. Please ensure that the time windows are ordered from earliest to latest and they do not overlap")
		}

	}

	return exceed24h, nil
}


func validateOptions(input *dto.OptimizationPostInput) (dto.OptimizationOptions, []string, error) {
	var warnings []string

	options := input.Options
	if len(options.Objective.TravelCost) != 0 {
		if options.Objective.TravelCost != "distance" && options.Objective.TravelCost != "duration" &&
			options.Objective.TravelCost != "customized" && options.Objective.TravelCost != "air_distance" {
			return options, warnings, fmt.Errorf("invalid value for \"travel_cost\" specified. Please ensure that the \"travel_cost\" belongs to the following options: \"distance\", \"duration\", \"air_distance\", or \"customized\"")
		}
		if options.Objective.TravelCost == "customized" {
			length := len(strings.Split(input.Locations.Location, "|"))
			err := validateCostMatrix(length, input.CostMatrix)
			if err != nil {
				return options, warnings, err
			}
		}
	} else {
		options.Objective.TravelCost = "duration"
	}

	routingOptions := options.Routing
	if routingOptions.Mode != nil {
		input.Mode = routingOptions.Mode
	} else if input.Mode != nil {
		routingOptions.Mode = input.Mode
	}

	if routingOptions.TruckSize != nil && len(*routingOptions.TruckSize) > 0 {
		truckSizes := strings.Split(*routingOptions.TruckSize, ",")
		if len(truckSizes) != 3 {
			return options, warnings, fmt.Errorf("the input for 'truck_size' is not in the correct format. Please ensure that 'truck_size' dimensions are specified as integer values")
		}
	}

	if (routingOptions.TruckSize != nil && len(*routingOptions.TruckSize) > 0) &&
		(routingOptions.Mode != nil && (*routingOptions.Mode == "4w" || *routingOptions.Mode == "car")) {
		warnings = append(warnings, "truck_size is ignored as mode=car")
	}

	if (routingOptions.TruckWeight != nil) &&
		(routingOptions.Mode != nil && (*routingOptions.Mode == "4w" || *routingOptions.Mode == "car")) {
		warnings = append(warnings, "truck_weight is ignored as mode=car")
	}

	return options, warnings, nil
}


func validateCostMatrix(length int, matrix [][]uint64) error {
	if len(matrix) != length {
		return fmt.Errorf("invalid length of cost matrix. Its size should be %d x %d", length, length)
	}
	for _, row := range matrix {
		if len(row) != length {
			return fmt.Errorf("invalid length of cost matrix. Its size should be %d x %d", length, length)
		}
	}
	return nil
}


func validateApproaches(locations dto.Locations) error {
	location := locations.Location
	length := len(strings.Split(location, "|"))
	if len(locations.Approaches) > 0 {
		approaches := locations.Approaches
		if len(approaches) != length {
			return fmt.Errorf("the number of approaches specified are not equal to the number of location coordinates provided in \"locations\" part. Please provide as many approaches as locations in the location array")
		}
		for _, approach := range approaches {
			if approach != "unrestricted" && approach != "curb" && approach != "" {
				return fmt.Errorf("the approach %s is invalid. Please ensure that the approach belongs to the following options: \"curb\", \"unrestricted\", or \"\" (empty string)", approach)
			}
		}
	}
	return nil
}
