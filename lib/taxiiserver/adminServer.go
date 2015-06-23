// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package taxiiserver

import (
	"log"
	"net/http"
	"net/url"
)

func (this *ServerType) AdminServerHandler(w http.ResponseWriter, r *http.Request) {
	if this.SysConfig.Logging.LogLevel >= 3 {
		log.Printf("DEBUG-3: Found Message on Admin Server Handler from %s", r.RemoteAddr)
	}

	urlValues, _ := url.ParseQuery(r.URL.RawQuery)

	// TODO look in to moving this to a JSON objet instead of URL parameters
	if val, ok := urlValues["reloadservices"]; ok {

		if val[0] == "true" {
			if this.SysConfig.Logging.LogLevel >= 3 {
				log.Println("DEBUG-3: Setting Reload Services to true via admin console")
			}

			this.ReloadServices = true
		}

	}
}
