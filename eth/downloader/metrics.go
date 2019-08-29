// Copyright 2015 The go-severeum Authors
// This file is part of the go-severeum library.
//
// The go-severeum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-severeum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-severeum library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/severeum/go-severeum/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("sev/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("sev/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("sev/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("sev/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("sev/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("sev/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("sev/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("sev/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("sev/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("sev/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("sev/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("sev/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("sev/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("sev/downloader/states/drop", nil)
)
