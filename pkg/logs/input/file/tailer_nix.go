// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

// +build !windows

package file

import (
	"io"
	"os"
	"path/filepath"

	"github.com/DataDog/datadog-agent/pkg/util/log"

	"fmt"

	"github.com/DataDog/datadog-agent/pkg/logs/decoder"
)

// setup sets up the file tailer
func (t *Tailer) setup(offset int64, whence int) error {
	fullpath, err := filepath.Abs(t.path)
	if err != nil {
		return err
	}
	t.tags = []string{fmt.Sprintf("filename:%s", filepath.Base(t.path))}
	log.Info("Opening ", t.path)
	f, err := os.Open(fullpath)
	if err != nil {
		return err
	}

	t.file = f
	ret, _ := f.Seek(offset, whence)
	t.readOffset = ret
	t.decodedOffset = ret

	return nil
}

// readForever lets the tailer tail the content of a file
// until it is closed or the tailer is stopped.
func (t *Tailer) readForever() {
	defer t.onStop()
	for {
		select {
		case <-t.stop:
			// stop reading data from file
			return
		default:
			// keep reading data from file
			inBuf := make([]byte, 4096)
			n, err := t.file.Read(inBuf)
			if err != nil && err != io.EOF {
				// an unexpected error occurred, stop the tailor
				t.source.Status.Error(err)
				log.Error("Err: ", err)
				return
			}
			if n == 0 {
				// wait for new data to come
				t.wait()
				continue
			}
			t.decoder.InputChan <- decoder.NewInput(inBuf[:n])
			t.incrementReadOffset(n)
		}
	}
}
