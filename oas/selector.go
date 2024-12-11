package oas

import (
	"net/http"
	"strings"
)

type APISelector func(r *http.Request) string

// HostBasedSelector returns an APISelector that selects an API based on the request host.
func HostBasedSelector(hostMap map[string]string) APISelector {
	return func(r *http.Request) string {
		return hostMap[r.Host]
	}
}

/* Ex:
   hostSelector := HostBasedSelector(map[string]string{
       "api.pets.com": "petstore",
       "api.users.com": "userapi",
   })
*/

// PathPrefixSelector returns an APISelector that selects an API based on the request path prefix.
func PathPrefixSelector(prefixMap map[string]string) APISelector {
	return func(r *http.Request) string {
		for prefix, apiName := range prefixMap {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return apiName
			}
		}
		return ""
	}
}

/* Ex:
   prefixSelector := PathPrefixSelector(map[string]string{
       "/v1": "petstore-v1",
       "/v2": "petstore-v2",
   })
*/

// HeaderSelector returns an APISelector that selects an API based on the value of a request header.
func HeaderSelector(headerName string) APISelector {
	return func(r *http.Request) string {
		return r.Header.Get(headerName)
	}
}

/* Ex:
   versionSelector := HeaderSelector("X-API-Version")
*/

// FixedSelector returns an APISelector that always selects the same API.
func FixedSelector(apiName string) APISelector {
	return func(r *http.Request) string {
		return apiName
	}
}
