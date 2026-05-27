// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// tarballDirectory walks dir and returns a gzipped tarball plus a sha256 of
// the file contents. node_modules, .git, and .pulumi are skipped (.pulumi can
// hold local stack state and credentials a user wouldn't expect to publish).
func tarballDirectory(dir string) ([]byte, string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, "", fmt.Errorf("stat %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, "", fmt.Errorf("%q is not a directory", dir)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hasher := sha256.New()
	entries := 0

	walkErr := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		base := d.Name()
		if base == "node_modules" || base == ".git" || base == ".pulumi" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		var link string
		if fi.Mode()&os.ModeSymlink != 0 {
			link, err = os.Readlink(p)
			if err != nil {
				return fmt.Errorf("read symlink %q: %w", rel, err)
			}
		}
		hdr, err := tar.FileInfoHeader(fi, link)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		entries++
		switch {
		case fi.Mode()&os.ModeSymlink != 0:
			// fold link target into the content hash so target swaps trigger drift
			if _, err := hasher.Write([]byte(link)); err != nil {
				return err
			}
		case fi.Mode().IsRegular():
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(io.MultiWriter(tw, hasher), f); err != nil {
				return err
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, "", walkErr
	}
	if err := tw.Close(); err != nil {
		return nil, "", err
	}
	if err := gz.Close(); err != nil {
		return nil, "", err
	}
	if entries == 0 {
		return nil, "", fmt.Errorf("%q is empty (all files were skipped)", dir)
	}
	return buf.Bytes(), hex.EncodeToString(hasher.Sum(nil)), nil
}
