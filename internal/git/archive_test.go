package git

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"gitlab.com/gitlab-org/gitlab-workhorse/internal/testhelper"
)

func TestParseBasename(t *testing.T) {
	for _, testCase := range []struct{ in, out string }{
		{"", "tar.gz"},
		{".tar.gz", "tar.gz"},
		{".tgz", "tar.gz"},
		{".gz", "tar.gz"},
		{".tar.bz2", "tar.bz2"},
		{".tbz", "tar.bz2"},
		{".tbz2", "tar.bz2"},
		{".tb2", "tar.bz2"},
		{".bz2", "tar.bz2"},
	} {
		basename := "archive" + testCase.in
		out, ok := parseBasename(basename)
		if !ok {
			t.Fatalf("parseBasename did not recognize %q", basename)
		}

		if out != testCase.out {
			t.Fatalf("expected %q, got %q", testCase.out, out)
		}
	}
}

func TestFinalizeArchive(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "gitlab-workhorse-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tempFile.Close()

	// Deliberately cause an EEXIST error: we know tempFile.Name() already exists
	err = finalizeCachedArchive(tempFile, tempFile.Name())
	if err != nil {
		t.Fatalf("expected nil from finalizeCachedArchive, received %v", err)
	}
}

func TestSetArchiveHeaders(t *testing.T) {
	for _, testCase := range []struct{ in, out string }{
		{"zip", "application/zip"},
		{"zippy", "application/octet-stream"},
		{"rezip", "application/octet-stream"},
		{"_anything_", "application/octet-stream"},
	} {
		w := httptest.NewRecorder()

		// These should be replaced, not appended to
		w.Header().Set("Content-Type", "test")
		w.Header().Set("Content-Length", "test")
		w.Header().Set("Content-Disposition", "test")
		w.Header().Set("Cache-Control", "test")

		setArchiveHeaders(w, testCase.in, "filename")

		testhelper.AssertResponseHeader(t, w, "Content-Type", testCase.out)
		testhelper.AssertResponseHeader(t, w, "Content-Length")
		testhelper.AssertResponseHeader(t, w, "Content-Disposition", `attachment; filename="filename"`)
		testhelper.AssertResponseHeader(t, w, "Cache-Control", "private")
	}
}
