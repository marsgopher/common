package concurrency

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
)

var (
	Web   = fakeSearch("web")
	Image = fakeSearch("image")
	Video = fakeSearch("video")
)

type Result string
type Search func(ctx context.Context, query string) (Result, error)

func fakeSearch(kind string) Search {
	return func(ctx context.Context, query string) (Result, error) {
		return Result(fmt.Sprintf("%s result for %q", kind, query)), nil
	}
}

// References:
// https://cs.opensource.google/go/x/sync/+/886fb937:errgroup/errgroup_test.go;l=38
// https://go.dev/blog/examples
func ExampleSemaErrGroup_justErrors() {
	g := NewSemaErrGroup(4)
	var urls = []string{
		"https://go.upyun.com",
		"https://www.upyun.com",
		"https://gitlab.s.upyun.com",
	}
	for i := range urls {
		url := urls[i]
		g.Do(func() error {
			// Fetch the URL
			resp, err := http.Get(url)
			if err == nil {
				_ = resp.Body.Close()
			}
			return err
		})
	}
	// Wait for all HTTP fetches to complete
	if err := g.Wait(); err == nil {
		fmt.Println("Successfully fetches all URLs.")
	}
}

func ExampleSemaErrGroup_parallels() {
	SearchAll := func(ctx context.Context, query string) ([]Result, error) {
		g, ctx := NewSemaErrGroupWithContext(ctx, 4)
		searches := []Search{Web, Image, Video}
		results := make([]Result, len(searches))
		for i, search := range searches {
			i, search := i, search // https://golang.org/doc/faq#closures_and_goroutines
			g.Do(func() error {
				result, err := search(ctx, query)
				if err == nil {
					results[i] = result
				}
				return err
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
		return results, nil
	}

	results, err := SearchAll(context.Background(), "goland")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}

	// Output:
	// web result for "goland"
	// image result for "goland"
	// video result for "goland"
}

func TestSemaErrGroupZeroGroup(t *testing.T) {
	err1 := errors.New("errgroup_test: 1")
	err2 := errors.New("errgroup_test: 2")

	cases := []struct {
		errs []error
	}{
		{errs: []error{}},
		{errs: []error{nil}},
		{errs: []error{err1}},
		{errs: []error{err1, nil}},
		{errs: []error{err1, nil, err2}},
	}

	for _, tc := range cases {
		g := NewSemaErrGroup(4)

		var firstErr error
		for i, err := range tc.errs {
			err := err
			g.Do(func() error { return err })

			if firstErr == nil && err != nil {
				firstErr = err
			}

			if gErr := g.Wait(); gErr != firstErr {
				t.Errorf("after %T.Do(func() error { return err }) for err in %v\n"+
					"g.Wait() = %v; want %v",
					g, tc.errs[:i+1], err, firstErr)
			}
		}
	}
}

func TestSemaErrGroupWithContext(t *testing.T) {
	errDoom := errors.New("group_test: doomed")

	cases := []struct {
		errs []error
		want error
	}{
		{want: nil},
		{errs: []error{nil}, want: nil},
		{errs: []error{errDoom}, want: errDoom},
		{errs: []error{errDoom, nil}, want: errDoom},
	}

	for _, tc := range cases {
		g, ctx := NewSemaErrGroupWithContext(context.Background(), 4)

		for _, err := range tc.errs {
			err := err
			g.Do(func() error { return err })
		}

		if err := g.Wait(); err != tc.want {
			t.Errorf("after %T.Do(func() error { return err }) for err in %v\n"+
				"g.Wait() = %v; want %v",
				g, tc.errs, err, tc.want)
		}

		canceled := false
		select {
		case <-ctx.Done():
			canceled = true
		default:
		}
		if !canceled {
			t.Errorf("after %T.Do(func() error { return err }) for err in %v\n"+
				"ctx.Done() was not closed",
				g, tc.errs)
		}
	}
}
