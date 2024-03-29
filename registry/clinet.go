package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"zkCache/zklog"
)

func RegisterService(r RegistrationVO) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(r)
	if err != nil {
		zklog.Logger.WithField("err", err.Error()).Error()
		return err
	}
	res, err := http.Post(
		ServiceURL,
		"application/json",
		buf)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to register service. Registry service "+
			"responded with code %v", res.StatusCode)
	}
	return nil
}

func ShutdownService(serviceName ServiceName, url string) error {
	r := RegistrationVO{
		ServiceName: serviceName,
		ServiceURL:  url,
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(r)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
		return err
	}

	req, err := http.NewRequest(
		http.MethodDelete,
		ServiceURL,
		buf,
	)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to deregister service. "+
			"Register service response with code %v", res.StatusCode)
	}

	return nil
}
