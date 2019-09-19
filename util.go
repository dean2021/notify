// Copyright 2018 CSOIO.COM, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 公用函数

package notify

import (
	"github.com/sirupsen/logrus"
	"strconv"
)

func bytes2str(b []byte) string {
	return string(b)
}

func BytesToInt64(buf []byte) int64 {
	value, err := strconv.ParseInt(bytes2str(buf), 10, 64)
	if err != nil {
		logrus.Error(err)
	}
	return value
}
