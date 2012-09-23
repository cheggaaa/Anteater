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

package aerpc

import (
	"strings"
)


const (
	DEFAULT_HOST = "127.0.0.1"
	DEFAULT_PORT = "32032"
	
	SERVER_FLAG_FORMAT = `
	Server addr format:
	default addr: `+DEFAULT_HOST+`:`+DEFAULT_PORT+`
	examples:
		-s=192.168.1.2 will be 192.168.1.2:` + DEFAULT_PORT+`
		-s=:32033 will be `+DEFAULT_HOST+`:32033
		-s=anteater.local:3234 will be anteater.local:3234
	`;
)

func NormalizeAddr(flagValue string) (addr string) {
	addrs := strings.Split(flagValue, ":")
	var host, port string
	if len(addrs) == 1 {
		host = addrs[0]
		port = DEFAULT_PORT
	} else if len(addrs) == 2 {
		host = addrs[0]
		port = addrs[1]
	}
	if len(host) == 0 {
		host = DEFAULT_HOST 
	}
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	addr = host + ":" + port
	return
}
