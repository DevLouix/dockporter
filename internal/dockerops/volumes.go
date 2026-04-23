package dockerops

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExportVolumeStream creates a streaming .tar.gz archive of a host directory.
func (dm *DockerManager) ExportVolumeStream(sourcePath string) (io.ReadCloser, <-chan error) {
	pr, pw := io.Pipe()
	errChan := make(chan error, 1)

	go func() {
		var walkErr error
		// ADVANCED: Ensure writers are closed BEFORE the pipe closes to flush gzip footers.
		gw := gzip.NewWriter(pw)
		tw := tar.NewWriter(gw)

		defer func() {
			tw.Close()
			gw.Close()
			if walkErr != nil {
				pw.CloseWithError(walkErr)
			} else {
				pw.Close()
			}
			errChan <- walkErr
		}()

		walkErr = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Clean relative paths
			relPath, err := filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}
			
			// Skip the root folder itself
			if relPath == "." {
				return nil
			}

			// ADVANCED: Handle symlinks properly during export
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}
			header.Name = relPath

			if info.Mode()&os.ModeSymlink != 0 {
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				header.Linkname = link
			}

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if info.Mode().IsRegular() {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				defer f.Close()
				_, copyErr := io.Copy(tw, f)
				return copyErr
			}
			return nil
		})
	}()

	return pr, errChan
}

// ImportVolumeFromStream extracts a streaming .tar.gz directly to disk (Zero intermediate files)
func (dm *DockerManager) ImportVolumeFromStream(stream io.Reader, destinationPath string) error {
	fmt.Printf("📥 Streaming volume unpack to: %s\n", destinationPath)

	cleanDest := filepath.Clean(destinationPath)
	if err := os.MkdirAll(cleanDest, 0755); err != nil {
		return err
	}

	gr, err := gzip.NewReader(stream)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// SECURITY FIX: Prevent Tar-Slip / Path Traversal attacks
		target := filepath.Join(cleanDest, filepath.Clean(header.Name))
		if !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("security violation: invalid file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			err := func() error {
				// Ensure parent directory exists before writing file
				os.MkdirAll(filepath.Dir(target), 0755)
				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
				if err != nil {
					return err
				}
				defer f.Close()

				_, err = io.Copy(f, tr)
				return err
			}()
			if err != nil {
				return err
			}
		case tar.TypeSymlink, tar.TypeLink:
			// ADVANCED: Safely restore symlinks
			os.MkdirAll(filepath.Dir(target), 0755)
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Ignore "file exists" errors for symlinks, or handle them based on strictness
				if !os.IsExist(err) {
					return err
				}
			}
		}
	}
	return nil
}