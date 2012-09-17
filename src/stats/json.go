package stats

import (
	"encoding/json"
	"utils"
)

type StatsJson struct {
	Anteater *Anteater `json:"anteater"`
	Storage *Storage   `json:"storage"`
	Allocate map[string]uint64 `json:"allocate"`
	Counters map[string]uint64 `json:"counters"`
	Traffic  map[string]uint64 `json:"traffic"`
	TrafficH  map[string]string `json:"trafficHuman"`
	Env *Env `json:"env"`
}


func (s *Stats) AsJson() (b []byte) {
	s.Refresh()
	sj := &StatsJson{
		Anteater : s.Anteater,
		Storage  : s.Storage,
		Env : s.Env, 
		Traffic  : make(map[string]uint64),
		TrafficH : make(map[string]string),
		Allocate : make(map[string]uint64),
		Counters : make(map[string]uint64),
	}
	
	sj.Allocate["append"] = s.Allocate.Append.GetValue()
	sj.Allocate["in"] = s.Allocate.In.GetValue()
	sj.Allocate["replace"] = s.Allocate.Replace.GetValue()
	
	sj.Counters["get"] = s.Counters.Get.GetValue()
	sj.Counters["add"] = s.Counters.Add.GetValue()
	sj.Counters["delete"] = s.Counters.Delete.GetValue()
	sj.Counters["notFound"] = s.Counters.NotFound.GetValue()
	sj.Counters["notModified"] = s.Counters.NotModified.GetValue()
	
	sj.Traffic["in"] = s.Traffic.Input.GetValue()
	sj.Traffic["out"] = s.Traffic.Output.GetValue()
	sj.TrafficH["in"] = utils.HumanBytes(int64(sj.Traffic["in"]))
	sj.TrafficH["out"] = utils.HumanBytes(int64(sj.Traffic["out"]))
	
	b, err := json.Marshal(sj)	
	if err != nil {
		panic(err)
	}
	return
}