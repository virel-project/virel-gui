//go:build !js
// +build !js

package lang

import "github.com/cloudfoundry/jibber_jabber"

func DetectIETF() (string, error) {
	return jibber_jabber.DetectIETF()
}
