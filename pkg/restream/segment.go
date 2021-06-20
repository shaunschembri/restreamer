package restream

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const decrypterBuffer = 32768

func (r *Restream) getSegments(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case segment := <-r.segments:
			switch segment.keyMethod {
			case "AES-128":
				r.decrypter = &aes128{
					iv:         segment.iv,
					keyURL:     segment.keyURL,
					bufferSize: decrypterBuffer,
					request: request{
						userAgent: r.UserAgent,
						client:    &http.Client{},
					},
				}

				if err := r.decrypter.init(ctx); err != nil {
					r.errors <- fmt.Errorf("error initiating decrypter %s: %w", r.decrypter.info(), err)
					return
				}

			case "NONE":
				r.decrypter = nil
			default:
				r.decrypter = nil
				r.errors <- fmt.Errorf("key method %s is not supported", segment.keyMethod)
				return
			}

			if err := r.writeSegment(ctx, segment.url); err != nil {
				r.errors <- err
				return
			}
		}
	}
}

func (r *Restream) writeSegment(ctx context.Context, url string) error {
	request := request{
		userAgent: r.UserAgent,
		client:    &http.Client{},
	}

	response, err := request.do(ctx, url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if r.Writer == nil {
		return fmt.Errorf("stopping streaming as writer is nil")
	}

	reader := bufio.NewReaderSize(response.Body, r.ReadBufferSize)
	startTime := time.Now()
	if _, err := reader.Peek(r.ReadBufferSize); err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("error filling buffer: %w", err)
		}
	}
	r.currentBandwidth = uint32(float64(reader.Buffered()*8) / time.Since(startTime).Seconds())

	segmentSize := 0
	writer := NewStreamWriter(ctx, r.Writer)
	buffer := make([]byte, decrypterBuffer)

	for {
		bytesRead, readErr := io.ReadFull(reader, buffer)
		if errors.Is(readErr, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stream: %w", err)
		}

		decryptedPayload, err := r.decrypt(buffer, bytesRead)
		if err != nil {
			return fmt.Errorf("cannot decrypt: %w", err)
		}

		bytesWritten, err := writer.Write(decryptedPayload)
		if err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}

		r.streamedBytes += int64(bytesWritten)
		segmentSize += bytesWritten
	}

	return nil
}

func (r *Restream) decrypt(payload []byte, bytesRead int) ([]byte, error) {
	if r.decrypter != nil {
		decryptedPayload, err := r.decrypter.decrypt(payload[:bytesRead])
		if err != nil {
			return nil, fmt.Errorf("[%s] %w", r.decrypter.info(), err)
		}

		return decryptedPayload, nil
	}

	return payload[:bytesRead], nil
}

type writerCtx struct {
	ctx    context.Context
	writer io.Writer
}

func (w *writerCtx) Write(p []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, fmt.Errorf("context error: %w", err)
	}

	n, err := w.writer.Write(p)
	if err != nil {
		return n, fmt.Errorf("write error: %w", err)
	}

	return n, nil
}

func NewStreamWriter(ctx context.Context, writer io.Writer) io.Writer {
	return &writerCtx{ctx: ctx, writer: writer}
}
