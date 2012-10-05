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

package amazon

import (
	"config"
	"time"
	"sync"
	"fmt"
	"net/http"
	"aelog"
	"strings"
)

const (
	INVALIDATION_URL = "https://cloudfront.amazonaws.com/2012-07-01/distribution/%s/invalidation"
	INVALIDATION_REQUEST = `<?xml version="1.0" encoding="UTF-8"?>
<InvalidationBatch xmlns="http://cloudfront.amazonaws.com/doc/2012-07-01/">
   <Paths>
      <Quantity>%d</Quantity>
      <Items>
         %s
      </Items>
   </Paths>
   <CallerReference>%d.%d</CallerReference>
</InvalidationBatch>
`;
	INVALIDATION_ITEM = "<Path>/%s</Path>";
)


func NewCloudFront(c *config.Config) (cf *CloudFront) {
	cf = &CloudFront{
		conf : c,
		changeList : make([]string, 0),
	}
	cf.Init()
	return
}

type CloudFront struct {
	conf *config.Config
	changeList []string
	m *sync.Mutex
}

func (cf *CloudFront) Init() {
	if cf.conf.AmazonCFEnable {
		cf.m = &sync.Mutex{}
		go func() {
			tick := time.Tick(cf.conf.AmazonInvalidationDuration)
			for _ = range tick {
				cf.Tick()
			}
		}()
	}
}

func (cf *CloudFront) OnChange(name string) {
	cf.m.Lock()
	defer cf.m.Unlock()
	cf.changeList = append(cf.changeList, name)
}

func (cf *CloudFront) SendInvalidate() {
	cf.m.Lock()
	
	if len(cf.changeList) == 0 {
		cf.m.Unlock()
		return
	}
	aelog.Debugf("Amazon. Send invalidation request for a %d files", len(cf.changeList))
	var items string	
	for _, name := range cf.changeList {		
		items += fmt.Sprintf(INVALIDATION_ITEM, name)
	}
	data := fmt.Sprintf(INVALIDATION_REQUEST, len(cf.changeList), items, time.Now().Unix(), time.Now().UnixNano())
	url := fmt.Sprintf(INVALIDATION_URL, cf.conf.AmazonCFDistributionId)
	cf.changeList = make([]string, 0)
	cf.m.Unlock()	
	cf.request(url, data)
}


func (cf *CloudFront) request(url string, data string) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		aelog.Warnf("Amazon. Inavlidation request has err: %v", err)
		return
	}
	defer req.Body.Close()
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("Authorization", cf.conf.AmazonCFAuthentication)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		aelog.Warnf("Amazon. Inavlidation response has err: %v", err)
		return
	}
	defer res.Body.Close()
	switch res.StatusCode {
		case 200, 201:
			return;
		default:
			aelog.Infof("Amazon. Inavlidation response return non 2XX status: %s", res.Status)
	}
}

func (cf *CloudFront) Tick() {
	cf.SendInvalidate()
}

func (cf *CloudFront) Close() {
	cf.SendInvalidate()
}