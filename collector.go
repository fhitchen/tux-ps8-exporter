package main

import (
	"bufio"
	"bytes"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var serviceLastTime = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "tuxedo_service_time_last", Help: "Displays the last time a specific TUXEDO service reported."},
	[]string{"Service", "Routine", "Program", "SRVID"})

var serviceMaxTime = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "tuxedo_service_time_max", Help: "Displays the MAX time a specific TUXEDO service reported."},
	[]string{"Service", "Routine", "Program", "SRVID"})

var serviceReqs = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "tuxedo_service_requests_total",
		Help: "How many TUXEDO requests processed, partitioned by status and service.",
	},
	[]string{"Service", "Routine", "Program", "SRVID", "status"},
)

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}
func getScale(s string) float64 {
	switch s {
	case "K":
		return 1000
	case "M":
		return 1000000
	default:
	}
	return 1
}
func getPower() {
	cmd := exec.Command("/bin/sh", "-c", "xadmin psc")

	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	// Output from the previous ipmi sensor: 52,PSU_Input_Power,204.00,W,'OK'
	// We are only interested in the watts, so we only store that part.
	//log.Printf("%s\n", out)
	/*
	   Nd Service Name Routine Name Prog Name SRVID #SUCC #FAIL      MAX     LAST STAT
	   -- ------------ ------------ --------- ----- ----- ----- -------- -------- -----
	    1 .TMIB        MIB          tpadmsv      10     0     0      0ms      0ms AVAIL
	    1 .TMIB-1-10   MIB          tpadmsv      10     0     0      0ms      0ms AVAIL
	    1 .TMIB-1      MIB          tpadmsv      10     0     0      0ms      0ms AVAIL
	    1 .TMIB        MIB          tpadmsv      11     0     0      0ms      0ms AVAIL
	    1 .TMIB-1-11   MIB          tpadmsv      11     0     0      0ms      0ms AVAIL
	    1 .TMIB-1      MIB          tpadmsv      11     0     0      0ms      0ms AVAIL
	    1 BALANCE      BALANCE      banksv        1   449     0     18ms      0ms AVAIL
	    1 BALANCE      BALANCE      banksv        2    2K     3     20ms      0ms AVAIL
	    1 csmUUsage    csmUUsage    csmUUsage     3     1     0      0ms      0ms AVAIL
	    1 csmUUsage    csmUUsage    csmUUsage     4     0     0      0ms      0ms AVAIL */
	a := regexp.MustCompile(`^\s?\d+\s+(?P<Service>\S+)\s+(?P<Routine>\S+)\s+(?P<Prog>\S+)\s+(?P<SRVID>\d+)\s+(?P<SUCC>\d+)(?P<S>\S?)\s+(?P<FAIL>\d+)(?P<F>\S?)\s+(?P<MAX>\d+)(?P<M>\S+)\s+(?P<LAST>\d+)(?P<L>\S+)\s+(?P<STAT>\S+)`)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		m := reSubMatchMap(a, scanner.Text())
		s, _ := strconv.ParseFloat(m["SUCC"], 64)
		serviceReqs.WithLabelValues(m["Service"], m["Routine"], m["Prog"], m["SRVID"], "SUCC").Set(s * getScale(m["S"]))
		s, _ = strconv.ParseFloat(m["FAIL"], 64)
		serviceReqs.WithLabelValues(m["Service"], m["Routine"], m["Prog"], m["SRVID"], "FAIL").Set(s)
		s, _ = strconv.ParseFloat(m["MAX"], 64)
		serviceMaxTime.WithLabelValues(m["Service"], m["Routine"], m["Prog"], m["SRVID"]).Set(s)
		s, _ = strconv.ParseFloat(m["LAST"], 64)
		serviceLastTime.WithLabelValues(m["Service"], m["Routine"], m["Prog"], m["SRVID"]).Set(s)
	}
}

// This is the first function to execute
func init() {
	prometheus.MustRegister(serviceMaxTime)
	prometheus.MustRegister(serviceLastTime)
	prometheus.MustRegister(serviceReqs)
	getPower()
}

func self_update() {
	rand.Seed(time.Now().Unix())
	for {
		getPower()
		time.Sleep(time.Second)
	}
}
