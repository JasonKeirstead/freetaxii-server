// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package taxiiserver

import (
	"github.com/freetaxii/freetaxii-server/lib/config"
)

// ----------------------------------------------------------------------
// Define Server Type
// ----------------------------------------------------------------------

type ServerType struct {
	SysConfig      *config.ServerConfigType
	ReloadServices bool
	CurrentTaxiiServicesType
}

// This type will hold the list of currently configured TAXII Services as found
// in the database
type CurrentTaxiiServicesType struct {
	CurrentTaxiiServices []TaxiiServiceType
}

type TaxiiServiceType struct {
	ServiceType string
	Available   bool
	Address     string
}
