/*
  Copyright 2012 Sergey Cherepanov (https://github.com/cheggaaa)

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package stats

import (
	"encoding/json"
	"utils"
)

type StatsInfo struct {
	Anteater *Anteater `json:"anteater"`
	Storage *Storage   `json:"storage"`
	Allocate map[string]uint64 `json:"allocate"`
	Counters map[string]uint64 `json:"counters"`
	Traffic  map[string]uint64 `json:"traffic"`
	TrafficH  map[string]string `json:"trafficHuman"`
	Env *Env `json:"env"`
}


func (s *Stats) AsJson() (b []byte) {
	sj := s.Info()
	b, err := json.Marshal(sj)	
	if err != nil {
		panic(err)
	}
	return
}

func (s *Stats) Info() *StatsInfo {
	s.Refresh()
	sj := &StatsInfo{
		Anteater : s.Anteater,
		Storage  : s.Storage,
		Env : s.Env, 
		Traffic  : map[string]uint64{"in":0, "out":0},
		TrafficH : map[string]string{"in":"0", "out":"0"},
		Allocate : map[string]uint64{"append":0, "in":0,"replace":0},
		Counters : map[string]uint64{"add":0, "get":0,"delete":0,"notFound":0,"notModified":0},
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
	return sj
}