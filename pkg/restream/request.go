package restream

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type request struct {
	client    *http.Client
	userAgent string
}

func (r request) do(ctx context.Context, requestURL string) (*http.Response, error) {
	response, err := r.attemptRequest(ctx, requestURL)
	if err != nil || response == nil {
		log.Printf("request to %s failed with error %v. Will retry in 1 second", requestURL, err)
		time.Sleep(time.Second)
		return r.do(ctx, requestURL)
	}

	if response.StatusCode < http.StatusBadRequest {
		return response, nil
	}

	if response.StatusCode != http.StatusNotFound {
		log.Printf("request to %s failed with status code %d. Will retry in 1 second", requestURL, response.StatusCode)
		time.Sleep(time.Second)
		return r.do(ctx, requestURL)
	}

	return nil, fmt.Errorf("request to %s failed with status code %d", requestURL, response.StatusCode)
}

func (r request) attemptRequest(context context.Context, requestURL string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(context, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	request.Header.Add("User-Agent", r.userAgent)
	request.Header.Add("Accept-Encoding", "gzip")

	response, err := r.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		response.Body = reader
	}

	return response, nil
}

func (r request) resolveReference(uri string, referenceURL *url.URL) (*url.URL, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("cannot parse segment uri %s: %w", uri, err)
	}

	return referenceURL.ResolveReference(parsedURI), nil
}
