package utils

import (
	"fmt"
	"strings"
	"time"
)

func ParsePeriodDateTime(timeQuery string, startDateTimeQuery string, endDateTimeQuery string) (time.Time, time.Time, error) {
	// There are 2 types of query params that can be used:
	// 1 - time: using the patterns: ?time=2019-01-02T00:00:00Z/2019-06-02T00:00:00Z | ?time=2019-01-02T00:00:00Z/ | ?time=/2019-06-02T00:00:00Z
	// 2 - startDateTime and endDateTime: ?startDateTime=2019-01-02T00:00:00Z&endDateTime=2019-06-02T00:00:00Z | ?startDateTime=2019-01-02T00:00:00Z | ?endDateTime=2019-06-02T00:00:00Z
	var startTime time.Time
	var endTime time.Time
	var err error

	timeQueryParamFormatErrorMessage := "The query param `time` must be in the one of the following patterns [ /" + time.RFC3339 + " | " + time.RFC3339 + "/ | " + time.RFC3339 + "/" + time.RFC3339 + " ] (Ex: \"/2019-06-02T00:00:00Z\", \"2019-06-02T00:00:00Z/\", \"2019-02-01T00:00:00/2019-06-02T00:00:00Z\")"
	if timeQuery != "" { // if the query param `time` was provided
		if strings.Contains(timeQuery, "/") {
			parts := strings.Split(timeQuery, "/")
			if len(parts) == 2 && parts[1] == "" { // pattern: "2019-06-02T00:00:00/Z"
				startTime, err = time.Parse(time.RFC3339, parts[0])
				endTime = time.Date(2200, time.January, 1, 0, 0, 0, 0, time.UTC)

			} else if len(parts) == 2 && parts[0] == "" { // pattern: "/2019-06-02T00:00:00Z"
				startTime = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
				endTime, err = time.Parse(time.RFC3339, parts[1])
				if err != nil {
					fmt.Printf("%s", err)
					return time.Time{}, time.Time{}, fmt.Errorf(timeQueryParamFormatErrorMessage)
				}

			} else if len(parts) == 2 && parts[0] != "" && parts[1] != "" { // pattern: "2019-02-02T00:00:00/2019-06-02T00:00:00Z"
				startTime, err = time.Parse(time.RFC3339, parts[0])
				if err != nil {
					fmt.Printf("%s", err)
					return time.Time{}, time.Time{}, fmt.Errorf(timeQueryParamFormatErrorMessage)
				}

				endTime, err = time.Parse(time.RFC3339, parts[1])
				if err != nil {
					fmt.Printf("%s", err)
					return time.Time{}, time.Time{}, fmt.Errorf(timeQueryParamFormatErrorMessage)
				}
			} else {
				fmt.Printf("%s", err)
				return time.Time{}, time.Time{}, fmt.Errorf(timeQueryParamFormatErrorMessage)
			}
		} else { // if the `/` was not provided
			fmt.Printf("%s", err)
			return time.Time{}, time.Time{}, fmt.Errorf(timeQueryParamFormatErrorMessage)
		}

	} else { // if startDateTime and/or endDateTime was provided

		if startDateTimeQuery == "" && endDateTimeQuery == "" {
			fmt.Printf("%s", err)
			return time.Time{}, time.Time{}, fmt.Errorf("You must provide either `time` query param or `startDateTime` and `endDateTime` query params to filter required data")
		}

		if startDateTimeQuery != "" {
			startTime, err = time.Parse(time.RFC3339, startDateTimeQuery)
			if err != nil {
				fmt.Printf("%s", err)
				return time.Time{}, time.Time{}, fmt.Errorf("The `startDateTime` query param must be in RFC3339 pattern, e.g. " + time.RFC3339)
			}
		}
		if endDateTimeQuery != "" {
			endTime, err = time.Parse(time.RFC3339, endDateTimeQuery)
			if err != nil {
				fmt.Printf("%s", err)
				return time.Time{}, time.Time{}, fmt.Errorf("The `endDateTime` query param must be in RFC3339 pattern, e.g. " + time.RFC3339)
			}
		}
	}
	if err != nil {
		fmt.Printf("%s", err)
		return time.Time{}, time.Time{}, fmt.Errorf("Error parsing request body. Err: %s", err)
	}

	return startTime, endTime, nil
}
