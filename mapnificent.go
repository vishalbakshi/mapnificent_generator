package main

import (
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/mapnificent/mapnificent_generator/mapnificent.pb"
	"github.com/nicolaspaton/gogtfs"
)

var (
	pathsString = flag.String("d", "", "Directories containing gtfs txt or zip files or zip file path (directories are traversed, multi coma separated: \"/here,/there\")")
	outputFile  = flag.String("o", "", "Output file")
	shouldLog   = flag.Bool("v", false, "Log to Stdout/err")
	needHelp    = flag.Bool("h", false, "Displays this help message...")
	feeds       map[string]*gtfs.Feed
)

func init() {
	feeds = make(map[string]*gtfs.Feed, 10)
}

const (
	HOUR_RANGE               = int32(3)
	IDENTICAL_STATION_RADIUS = 30
	WALK_STATION_RADIUS      = 150
)

func discoverGtfsPaths(path string) (results []string) {
	log.Println("discoverGtfsPaths: " + path)
	path = filepath.Clean(path)
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return
	}
	if fileInfo.IsDir() {
		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return
		}
		requiredFiles := gtfs.RequiredFiles
		requiredCalendarFiles := gtfs.RequiredEitherCalendarFiles
		foundCalendar := false
		foundFiles := make([]string, 0, len(requiredFiles))
		for _, fi := range fileInfos {
			name := fi.Name()
			if fi.IsDir() {
				subdirectoryResults := discoverGtfsPaths(path + "/" + name)
				for _, newpath := range subdirectoryResults {
					results = append(results, newpath)
				}
			} else if filepath.Ext(name) == ".zip" {
				results = append(results, path+"/"+name)
			} else {
				for _, f := range requiredFiles {
					if name == f { // This loops a little too much but hey...
						foundFiles = append(foundFiles, f)
					}
				}
				if !foundCalendar {
					for _, f := range requiredCalendarFiles {
						if name == f {
							foundCalendar = true
						}
					}
				}

			}
		}
		if len(foundFiles) == len(requiredFiles) && foundCalendar {
			results = append(results, path)
		}

	} else {
		if filepath.Ext(path) == ".zip" {
			results = append(results, path)
		}
	}
	return
}

func GetNetwork(feeds map[string]*gtfs.Feed) *mapnificent.MapnificentNetwork {

	network := new(mapnificent.MapnificentNetwork)

	var name string
	feedNr := 0

	for path, feed := range feeds {

		feedNr += 1

		stopWalked := make(map[uint]bool)

		if name == "" {
			name = getNameFromPath(path)
			// Store first feed path as name
			network.Cityid = proto.String(name)
		}

		stationMap := make(map[string]uint)
		lineMap := make(map[string]*list.List)

		for _, trip := range feed.Trips {
			tripHash := GetTripHash(trip)
			_, ok := lineMap[tripHash]
			if !ok {
				lineMap[tripHash] = list.New()
			}
			lineMap[tripHash].PushBack(trip)
		}

		for _, li := range lineMap {

			trip := li.Front().Value.(*gtfs.Trip)

			mapnificent_line := mapnificent.MapnificentNetwork_Line{
				LineId: proto.String(trip.Id + "|" + trip.Route.Id),
			}

			GetFrequencies(feed, li, &mapnificent_line)

			if len(mapnificent_line.LineTimes) == 0 {
				continue
			}

			network.Lines = append(network.Lines, &mapnificent_line)

			var lastStopArrival, lastStopDeparture uint
			var lastStop *mapnificent.MapnificentNetwork_Stop

			for _, stoptime := range trip.StopTimes {
				stopIndex := GetOrCreateMapnificentStop(feed, feedNr, stoptime.Stop, network, stationMap)
				mapnificentStop := network.Stops[stopIndex]

				_, walkedOk := stopWalked[stopIndex]
				if !walkedOk {
					// Search 500 m radius
					walkStopDistances := feed.StopCollection.StopDistancesByProximity(stoptime.Stop.Lat, stoptime.Stop.Lon, WALK_STATION_RADIUS)
					sameStopWalked := make(map[uint]bool)
					for _, walkStopDistance := range walkStopDistances {
						if walkStopDistance.Stop.Id == stoptime.Stop.Id {
							continue
						}

						walkStopIndex := GetOrCreateMapnificentStop(feed, feedNr, walkStopDistance.Stop, network, stationMap)
						if walkStopIndex == stopIndex {
							continue
						}
						_, sameStopWalkedOk := sameStopWalked[walkStopIndex]
						if sameStopWalkedOk {
							continue
						}
						sameStopWalked[walkStopIndex] = true

						walkTravelOption := new(mapnificent.MapnificentNetwork_Stop_TravelOption)
						walkTravelOption.Stop = proto.Int32(int32(walkStopIndex))
						walkTravelOption.WalkDistance = proto.Int32(int32(walkStopDistance.Distance))
						mapnificentStop.TravelOptions = append(mapnificentStop.TravelOptions, walkTravelOption)
					}
					stopWalked[stopIndex] = true
				}

				if lastStop != nil {
					delta := stoptime.ArrivalTime - lastStopDeparture
					stayDelta := lastStopDeparture - lastStopArrival
					travelOption := new(mapnificent.MapnificentNetwork_Stop_TravelOption)
					travelOption.Stop = proto.Int32(int32(stopIndex))
					travelOption.TravelTime = proto.Int32(int32(delta))
					travelOption.StayTime = proto.Int32(int32(stayDelta))
					travelOption.Line = mapnificent_line.LineId
					lastStop.TravelOptions = append(lastStop.TravelOptions, travelOption)
				}
				// log.Println("Travel options count", stopIndex, len(mapnificentStop.TravelOptions))
				lastStopArrival = stoptime.ArrivalTime
				lastStopDeparture = stoptime.DepartureTime
				lastStop = mapnificentStop
			}
		}
	}
	// for i, stop := range network.Stops {
	//   log.Println("check stops", i, len(stop.TravelOptions))
	// }
	return network
}

func GetOrCreateMapnificentStop(feed *gtfs.Feed, feedNr int, stop *gtfs.Stop,
	network *mapnificent.MapnificentNetwork,
	stationMap map[string]uint) uint {
	stationName := fmt.Sprintf("%d_%s", feedNr, stop.Id)
	stopIndex, ok := stationMap[stationName]
	if !ok {
		// Consider all stops in 50 meter radius as identical
		nearbyStops := feed.StopCollection.StopsByProximity(stop.Lat, stop.Lon, IDENTICAL_STATION_RADIUS)
		foundStopIndex := -1
		for _, nearbyStop := range nearbyStops {
			nearbyStopnName := fmt.Sprintf("%d_%s", feedNr, nearbyStop.Id)
			nearbystopIndex, ok := stationMap[nearbyStopnName]
			if ok {
				foundStopIndex = int(nearbystopIndex)
				break
			}
		}
		if foundStopIndex == -1 {
			mapnificentStop := new(mapnificent.MapnificentNetwork_Stop)
			mapnificentStop.Latitude = proto.Float64(stop.Lat)
			mapnificentStop.Longitude = proto.Float64(stop.Lon)
			network.Stops = append(network.Stops, mapnificentStop)
			stopIndex = uint(len(network.Stops) - 1)
		} else {
			stopIndex = uint(foundStopIndex)
		}
		stationMap[stationName] = stopIndex
	}
	return stopIndex
}

func b2i(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func GetWeekdaysForTrip(feed *gtfs.Feed, serviceId string) (weekdays int32) {
	weekdays = 0
	if calendar, ok := feed.Calendars[serviceId]; ok {
		weekdays = weekdays | (b2i(calendar.Monday) << 0)
		weekdays = weekdays | (b2i(calendar.Tuesday) << 1)
		weekdays = weekdays | (b2i(calendar.Wednesday) << 2)
		weekdays = weekdays | (b2i(calendar.Thursday) << 3)
		weekdays = weekdays | (b2i(calendar.Friday) << 4)
		weekdays = weekdays | (b2i(calendar.Saturday) << 5)
		weekdays = weekdays | (b2i(calendar.Sunday) << 6)
	}
	return weekdays
}

func round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

func GetFrequencies(feed *gtfs.Feed, trips *list.List, line *mapnificent.MapnificentNetwork_Line) {
	// Bitmask: 7 bits, Monday lowest bit
	// 31 Weekdays
	// 96 Weekends
	// 64 Sunday
	// 63 all but Sunday
	// 32 Saturday
	// 48 Friday/Saturday
	//

	service_ranges := []int32{
		1, 6, // Monday from 6
		// 96, 9, // Weekend from 9
		48, 21, // Friday Saturday evening
	}

	cache_weekdays := make(map[string]int32)
	service_trips := make(map[int32]*list.List)

	// trip := trips.Front().Value.(*gtfs.Trip)
	for i := 0; i < len(service_ranges); i += 2 {
		service_day := service_ranges[i]

		for trip := trips.Front(); trip != nil; trip = trip.Next() {
			realTrip := trip.Value.(*gtfs.Trip)
			weekdays, wd_ok := cache_weekdays[realTrip.ServiceId]
			if !wd_ok {
				weekdays = GetWeekdaysForTrip(feed, realTrip.ServiceId)
				cache_weekdays[realTrip.ServiceId] = weekdays
			}
			if service_day&weekdays >= service_day {
				_, ok := service_trips[service_day]
				if !ok {
					service_trips[service_day] = list.New()
				}
				service_trips[service_day].PushBack(realTrip)
			}
		}
		st, ok := service_trips[service_day]
		if ok {
			log.Println("Found count for service day", st.Len(), service_day)
		} else {
			log.Println("Found count for service day 0", service_day)
		}
	}

	for i := 0; i < len(service_ranges); i += 2 {
		wd := service_ranges[i]
		hour := service_ranges[i+1]
		tripList, ok := service_trips[wd]

		if !ok {
			// Found no trips for this service id
			continue
		}

		if wd == 0 {
			// Runs on no days (possibly only calendardates that we are ignording)
			continue
		}

		frequencyCounter := 0
		//
		// for trip := tripList.Front(); trip != nil; trip = trip.Next() {
		//
		//   realTrip := trip.Value.(*gtfs.Trip)
		//
		//   if len(realTrip.Frequencies) > 0 {
		//       frequencyCounter += 1
		//       for _, freq := range realTrip.Frequencies {
		//         start := math.Floor(float64(freq.StartTime) / float64(hour_range))
		//         stop := math.Floor(float64(freq.EndTime) / float64(hour_range))
		//         headway := round(float64(freq.HeadwaySecs) / 60.0)
		//     		mapnificent_line_time = mapnificent.MapnificentNetwork_Line_LineTime{
		// 					Interval: proto.Int32(int32(headway)),
		// 					Start:    proto.Int32(int32(start)),
		// 					Stop:     proto.Int32(int32(stop)),
		// 					Weekday:  proto.Int32(int32(wd)),
		// 				}
		// 				line.LineTimes = append(line.LineTimes, &mapnificent_line_time)
		//       }
		//   }
		// }
		//
		// if frequencyCounter == tripList.Len() {
		//   // All trips here have frequencies, no point in calculating
		//   continue
		// }

		depTimes := make([]int, tripList.Len())
		depTimesCounter := 0

		var lastTrip *gtfs.Trip

		for trip := tripList.Front(); trip != nil; trip = trip.Next() {

			lastTrip = trip.Value.(*gtfs.Trip)

			if len(lastTrip.Frequencies) > 0 {
				frequencyCounter += 1
				continue
			}

			if len(lastTrip.StopTimes) == 0 {
				continue
			}

			depTime := lastTrip.StopTimes[0].DepartureTime
			depHour := int32(depTime / (60 * 60))
			if depHour > hour && depHour <= (hour+HOUR_RANGE) {
				depTimes[depTimesCounter] = int(depTime)
				depTimesCounter += 1
			}
		}

		if depTimesCounter == 0 {
			continue
		}

		depTimes = depTimes[:depTimesCounter]

		averageInterval := -1

		if len(depTimes) > 1 {
			sort.Ints(depTimes)

			lastDep := -1
			intervalSum := 0.0
			for i := 0; i < len(depTimes); i++ {
				if lastDep != -1 {
					intervalSum += float64(depTimes[i] - lastDep)
				}
				lastDep = depTimes[i]
			}

			averageInterval = round(intervalSum / float64(len(depTimes)-1))
		} else if len(depTimes) == 1 {
			// only once in three hours? no pattern
			averageInterval = -1
		} else {
			averageInterval = -1
		}

		log.Println("Calculated headway for day",
			lastTrip.Route.ShortName, wd, hour, averageInterval/60.0)
		if averageInterval < 0 {
			log.Println("Details", len(depTimes))
		}

		if averageInterval > 0 {
			// log.Println("Appending", wd, hour, averageInterval, "for", line.LineId)
			mapnificent_line_time := NewLineTime(wd, hour, int32(averageInterval))
			line.LineTimes = append(line.LineTimes, &mapnificent_line_time)
		}
	}
}

func NewLineTime(wd int32, hour int32, interval int32) mapnificent.MapnificentNetwork_Line_LineTime {
	return mapnificent.MapnificentNetwork_Line_LineTime{
		Interval: proto.Int32(interval),
		Start:    proto.Int32(hour),
		Stop:     proto.Int32(hour + HOUR_RANGE),
		Weekday:  proto.Int32(wd),
	}
}

func GetTripHash(trip *gtfs.Trip) string {
	/* Gets a hash based on route and the actual trip stops */
	h := md5.New()
	// log.Println("Route", trip.Route.Id)
	io.WriteString(h, trip.Route.Id)
	io.WriteString(h, "||")
	io.WriteString(h, trip.Headsign)
	io.WriteString(h, "||")
	io.WriteString(h, hex.EncodeToString([]byte{trip.Direction}))
	// for _, stoptime := range trip.StopTimes {
	// 	io.WriteString(h, stoptime.Stop.Id)
	// 	io.WriteString(h, "||")
	// 	// log.Println("Stop", stoptime.Stop.Id)
	// }
	return hex.EncodeToString(h.Sum(nil)[:])
}

func getNameFromPath(path string) string {
	pathParts := strings.Split(path, "/")
	return pathParts[len(pathParts)-1]
}

func main() {
	flag.Parse()

	if *needHelp {
		flag.Usage()
		os.Exit(0)
	}

	pathsAll := strings.Split(*pathsString, ",")
	paths := make([]string, 0, len(pathsAll))

	for _, path := range pathsAll {
		if path != "" {
			// Why do I have to do this
			path, _ = filepath.Abs(path)
			newpaths := discoverGtfsPaths(path)
			if len(newpaths) > 0 {
				for _, p := range newpaths {
					paths = append(paths, p)
				}
			}
		}
	}

	if !*shouldLog {
		devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0) // Shouldn't be an error
		defer devNull.Close()                                 // Useless, is it not?
		log.SetOutput(devNull)
	}

	log.SetPrefix("gtfs - ")

	channels := make([]chan bool, 0, len(paths))
	totalStopTimes := 0
	if len(paths) > 0 {
		for _, path := range paths[:] {
			log.Println(path)
			if path != "" {
				channel := make(chan bool)
				channels = append(channels, channel)
				go func(path string, ch chan bool) {
					log.Println("Started loading", path)
					feed, err := gtfs.NewFeed(path)
					if err != nil {
						log.Fatal(err)
					} else {
						feed.Load()
						feeds[path] = feed
					}
					totalStopTimes = totalStopTimes + feed.StopTimesCount
					channel <- true
				}(path, channel)
			}
		}
	} else {
		log.Println("No Paths found")
	}

	// Waiting for jobs to finnish
	for _, c := range channels {
		<-c
	}

	absOutFile, _ := filepath.Abs(*outputFile)
	outfile, err := os.Create(absOutFile)
	if err != nil {
		log.Println("Error creating", absOutFile)
		return
	}
	log.Println("Getting Network")
	network := GetNetwork(feeds)

	log.Println("Marshalling...")
	bytes, err := proto.Marshal(network)
	outfile.Write(bytes)
	log.Println("Marshalling Done.")
	outfile.Close()
}
