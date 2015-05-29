// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package headers

import (
	"fmt"
	"github.com/freetaxii/libtaxii/defs"
	"net/http"
)

type HttpHeaderType struct {
	DebugLevel int
}

// --------------------------------------------------
// Verify HTTP Headers
// --------------------------------------------------

func (this *HttpHeaderType) VerifyHttpTaxiiHeaderValues(r *http.Request) error {

	// --------------------------------------------------
	// Version of the TAXII specification they are using
	// --------------------------------------------------
	if r.Header["X-Taxii-Services"] == nil {
		return fmt.Errorf("%s, TAXII Service Not Defined in HTTP Header X-Taxii-Services", r.RemoteAddr)
	}

	if r.Header["X-Taxii-Services"][0] != defs.TAXII_VERSION {
		return fmt.Errorf("%s, Unsupported TAXII Service, %s", r.RemoteAddr, r.Header["X-Taxii-Services"][0])
	}

	// --------------------------------------------------
	// TAXII message format the client wants in return, JSON, XML, etc.
	// --------------------------------------------------
	if r.Header["X-Taxii-Accept"] == nil {
		return fmt.Errorf("%s, Requested Encoding Not Defined in HTTP Header X-Taxii-Accept", r.RemoteAddr)
	}

	if r.Header["X-Taxii-Accept"][0] != defs.TAXII_MESSAGE_JSON {
		return fmt.Errorf("%s, Client Requested Response Encoding in X-Taxii-Accept is Unsupported, %s", r.RemoteAddr, r.Header["X-Taxii-Accept"][0])
	}

	// --------------------------------------------------
	// TAXII message format the client used on this message, JSON, XML, etc.
	// --------------------------------------------------
	if r.Header["X-Taxii-Content-Type"] == nil {
		return fmt.Errorf("%s, Supplied Content Encoding Not Defined in HTTP Header X-Taxii-Content-Type", r.RemoteAddr)
	}

	if r.Header["X-Taxii-Content-Type"][0] != defs.TAXII_MESSAGE_JSON {
		return fmt.Errorf("%s, Supplied Message Encoding in X-Taxii-Content-Type Is Unsupported, %s", r.RemoteAddr, r.Header["X-Taxii-Content-Type"][0])
	}

	return nil
}

// --------------------------------------------------
// Debug HTTP Headers
// --------------------------------------------------

func (this *HttpHeaderType) DebugHttpRequest(r *http.Request) {

	fmt.Println("DEBUG: --------------- BEGIN HTTP DUMP ---------------")
	fmt.Println("DEBUG: Method", r.Method)
	fmt.Println("DEBUG: URL", r.URL)
	fmt.Println("DEBUG: Proto", r.Proto)
	fmt.Println("DEBUG: ProtoMajor", r.ProtoMajor)
	fmt.Println("DEBUG: ProtoMinor", r.ProtoMinor)
	fmt.Println("DEBUG: Header", r.Header)
	fmt.Println("DEBUG: Body", r.Body)
	fmt.Println("DEBUG: ContentLength", r.ContentLength)
	fmt.Println("DEBUG: TransferEncoding", r.TransferEncoding)
	fmt.Println("DEBUG: Close", r.Close)
	fmt.Println("DEBUG: Host", r.Host)
	fmt.Println("DEBUG: Form", r.Form)
	fmt.Println("DEBUG: PostForm", r.PostForm)
	fmt.Println("DEBUG: MultipartForm", r.MultipartForm)
	fmt.Println("DEBUG: Trailer", r.Trailer)
	fmt.Println("DEBUG: RemoteAddr", r.RemoteAddr)
	fmt.Println("DEBUG: RequestURI", r.RequestURI)
	fmt.Println("DEBUG: TLS", r.TLS)
	fmt.Println("DEBUG: --------------- END HTTP DUMP ---------------")
	fmt.Println("\n")
	fmt.Println("DEBUG: --------------- BEGIN HEADER DUMP ---------------")
	for k, v := range r.Header {
		fmt.Println("DEBUG:", k, v)
	}
	fmt.Println("DEBUG: --------------- END HEADER DUMP ---------------")
}
