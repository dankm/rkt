// Copyright 2015 The rkt Authors
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

package image

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// downloadSession is an interface used by downloader for controlling
// the downloading process.
type downloadSession interface {
	// GetClient returns a client used for handling the
	// requests. It is a good place for e.g. setting up a redirect
	// handling.
	GetClient() (*http.Client, error)
	// GetRequest returns an HTTP request. Is is a good place to
	// add some headers to a request.
	GetRequest(u *url.URL) (*http.Request, error)
	// HandleStatus is mostly used to check if the response has
	// the required HTTP status. When this function returns either
	// an error or a false value, then the downloader will skip
	// getting contents of the response body.
	HandleStatus(res *http.Response) (bool, error)
	// GetBodyReader returns a reader used to get contents of the
	// response body.
	GetBodyReader(*http.Response) (io.Reader, error)
}

// downloader has a rather obvious purpose - it downloads stuff.
type downloader struct {
	// Session controls the download process
	Session downloadSession
}

// Download tries to fetch the passed URL and write the contents into
// a given writeSyncer instance.
func (d *downloader) Download(u *url.URL, out writeSyncer) error {
	d.ensureSession()
	client, err := d.Session.GetClient()
	if err != nil {
		return err
	}
	req, err := d.Session.GetRequest(u)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if stopNow, err := d.Session.HandleStatus(res); stopNow || err != nil {
		return err
	}

	reader, err := d.Session.GetBodyReader(res)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, reader); err != nil {
		return fmt.Errorf("failed to download %q: %v", u.String(), err)
	}

	if err := out.Sync(); err != nil {
		return fmt.Errorf("failed to sync data from %q to disk: %v", u.String(), err)
	}

	return nil
}

func (d *downloader) ensureSession() {
	if isReallyNil(d.Session) {
		d.Session = &defaultDownloadSession{}
	}
}

// default DownloadSession is very simple implementation of
// downloadSession interface. It returns a default client, a GET
// request without additional headers and treats any HTTP status of a
// response other than 200 as an error.
type defaultDownloadSession struct{}

func (s *defaultDownloadSession) GetBodyReader(res *http.Response) (io.Reader, error) {
	return res.Body, nil
}

func (s *defaultDownloadSession) GetClient() (*http.Client, error) {
	return http.DefaultClient, nil
}

func (s *defaultDownloadSession) GetRequest(u *url.URL) (*http.Request, error) {
	req := &http.Request{
		Method:     "GET",
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}
	return req, nil
}

func (s *defaultDownloadSession) HandleStatus(res *http.Response) (bool, error) {
	switch res.StatusCode {
	case http.StatusOK:
		return false, nil
	default:
		return false, fmt.Errorf("bad HTTP status code: %d", res.StatusCode)
	}
}
