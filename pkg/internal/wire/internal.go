package wire

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ------------------------------------------------------------------------------------
// General concurrency & error-handling logic
// ------------------------------------------------------------------------------------

func (w *Wire[T]) isClosedSafe() bool {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	return w.isClosed
}

func (w *Wire[T]) setClosedSafe() {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	w.isClosed = true
}

func (w *Wire[T]) attemptRecovery(elem T, originalElem T, originalErr error) (T, error) {
	var retryErr error
	for attempt := 1; attempt <= w.retryThreshold; attempt++ {
		retryElem, retryErr := w.insulatorFunc(w.ctx, elem, originalErr)
		if retryErr == nil {
			w.notifyInsulatorAttemptSuccess(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
			return retryElem, nil
		}
		if attempt == w.retryThreshold {
			if w.CircuitBreaker != nil {
				w.CircuitBreaker.RecordError()
			}
			w.notifyInsulatorFinalRetryFailure(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
			return retryElem, fmt.Errorf("retry threshold of %d reached with error: %v", w.retryThreshold, retryErr)
		}
		w.notifyInsulatorAttempt(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
		elem = retryElem
		time.Sleep(w.retryInterval)
	}
	return elem, retryErr
}

func (w *Wire[T]) startCircuitBreakerTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.cbLock.Lock()
			allowed := true
			if w.CircuitBreaker != nil {
				allowed = w.CircuitBreaker.Allow()
			}
			w.cbLock.Unlock()

			w.closeLock.Lock()
			if !w.isClosed {
				select {
				case w.controlChan <- allowed:
				default:
				}
			}
			w.closeLock.Unlock()
		}
	}
}

func (w *Wire[T]) encodeElement(elem T) {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()
	if w.encoder != nil {
		if encodeErr := w.encoder.Encode(w.OutputBuffer, elem); encodeErr != nil {
			w.handleError(elem, encodeErr)
		}
	}
}

func (w *Wire[T]) handleCircuitBreakerTrip(ctx context.Context, elem T) error {
	groundWires := w.CircuitBreaker.GetNeutralWires()
	if len(groundWires) > 0 {
		for _, gw := range groundWires {
			if err := gw.Submit(ctx, elem); err != nil {
				continue
			}
			w.notifyNeutralWireSubmission(elem)
		}
		return nil
	}
	w.notifyCircuitBreakerDropElement(elem)
	return nil
}

func (w *Wire[T]) handleError(elem T, err error) {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	if w.isClosed || w.errorChan == nil {
		return
	}
	w.errorChan <- types.ElementError[T]{Err: err, Elem: elem}
}

func (w *Wire[T]) handleErrorElements() {
	defer w.wg.Done()
	for {
		select {
		case e, ok := <-w.errorChan:
			if !ok {
				return
			}
			w.notifyElementTransformError(e.Elem, e.Err)
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Wire[T]) handleProcessingError(elem T, originalElem T, originalErr error) (T, error) {
	if w.insulatorFunc != nil && (w.CircuitBreaker == nil || w.CircuitBreaker.Allow()) {
		return w.attemptRecovery(elem, originalElem, originalErr)
	}
	if w.CircuitBreaker != nil {
		w.CircuitBreaker.RecordError()
	}
	return elem, originalErr
}

func (w *Wire[T]) transformElement(elem T) (T, error) {
	var err error
	originalElem := elem
	for _, transform := range w.transformations {
		elem, err = transform(elem)
		if err != nil {
			return w.handleProcessingError(elem, originalElem, err)
		}
	}
	if w.sensors != nil {
		w.notifyElementProcessed(elem)
	}
	return elem, nil
}

func (w *Wire[T]) transformElements() {
	for i := 0; i < int(w.maxBufferSize); i++ {
		w.wg.Add(1)
		go w.processChannel()
	}
}

func (w *Wire[T]) processChannel() {
	defer w.wg.Done()
	for {
		select {
		case elem, ok := <-w.inChan:
			if !ok {
				return
			}
			w.concurrencySem <- struct{}{}
			go func(e T) {
				defer func() { <-w.concurrencySem }()
				if w.isClosedSafe() {
					return
				}
				w.processElementOrDivert(e)
			}(elem)
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Wire[T]) processElementOrDivert(elem T) {
	if w.isClosedSafe() {
		return
	}
	processedElem, err := w.transformElement(elem)
	if w.shouldDivertError(err) {
		w.divertToGroundWires(elem)
		return
	}
	if err != nil {
		w.handleError(elem, err)
		return
	}

	// If encryption is enabled, do it now
	if w.EncryptOptions != nil && w.EncryptOptions.Enabled {
		encrypted, encErr := w.encryptItem(processedElem)
		if encErr != nil {
			w.handleError(processedElem, encErr)
			return
		}
		processedElem = encrypted
	}
	w.submitProcessedElement(processedElem)
}

func (w *Wire[T]) shouldDivertError(err error) bool {
	return err != nil && w.CircuitBreaker != nil && !w.CircuitBreaker.Allow()
}

func (w *Wire[T]) divertToGroundWires(elem T) {
	if len(w.CircuitBreaker.GetNeutralWires()) > 0 {
		for _, gw := range w.CircuitBreaker.GetNeutralWires() {
			w.notifyNeutralWireSubmission(elem)
			gw.Submit(w.ctx, elem)
		}
	} else {
		w.handleError(elem, fmt.Errorf("circuit breaker tripped but no ground wires to handle"))
	}
}

func (w *Wire[T]) submitProcessedElement(processedElem T) {
	w.encodeElement(processedElem)
	if w.OutputChan == nil {
		return
	}
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	if w.isClosed {
		return
	}
	select {
	case w.OutputChan <- processedElem:
	case <-w.ctx.Done():
		w.notifyCancel(processedElem)
	}
}

func (w *Wire[T]) processResisterElements(ctx context.Context) {
	var ticker *time.Ticker
	if w.surgeProtector == nil || !w.surgeProtector.IsBeingRateLimited() {
		ticker = time.NewTicker(1 * time.Second)
	} else {
		_, fillrate, _, _ := w.surgeProtector.GetRateLimit()
		if fillrate > 0 {
			ticker = time.NewTicker(fillrate)
		} else {
			ticker = time.NewTicker(1 * time.Second)
		}
	}
	defer ticker.Stop()
	w.processItems(ctx, ticker)
}

func (w *Wire[T]) processItems(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.surgeProtector != nil && w.surgeProtector.TryTake() {
				if item, err := w.surgeProtector.Dequeue(); err == nil {
					w.Submit(ctx, item.Data)
				} else {
					if w.shouldStopProcessing() {
						return
					}
				}
			}
		}
	}
}

func (w *Wire[T]) shouldStopProcessing() bool {
	return w.surgeProtector.GetResisterQueue() == 0
}

func (w *Wire[T]) submitNormally(ctx context.Context, elem T) error {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	if w.isClosed {
		return fmt.Errorf("cannot submit: input channel is closed")
	}
	select {
	case w.inChan <- elem:
		w.notifySubmit(elem)
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// -----------------------------------------------------------------------------
// AES-GCM Encryption & Decryption (GOB-based), No Type Checking of T
// -----------------------------------------------------------------------------

// decryptItem expects the inbound data to be base64 ciphertext if DecryptOptions.Enabled is true.
// Steps:
//  1. base64 decode inbound => raw ciphertext
//  2. AES-GCM decrypt => plaintext bytes
//  3. GOB decode => final domain object T
//
// If DecryptOptions is disabled, we just return the inbound data as-is.
func (w *Wire[T]) decryptItem(inbound T) (T, error) {
	if w.DecryptOptions == nil || !w.DecryptOptions.Enabled {
		// No decrypt => pass data through
		return inbound, nil
	}

	// 1) interpret inbound T as base64 ciphertext
	base64Str, ok := any(inbound).(string)
	if !ok {
		// The user gave a non-string while decrypt is on => config error
		return inbound, fmt.Errorf("wire decrypt error: inbound must be base64 string, got %T instead", inbound)
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return inbound, fmt.Errorf("wire decrypt error: invalid base64 ciphertext: %w", err)
	}

	// 2) AES-GCM decrypt
	plaintextBytes, err := decryptData(cipherBytes, w.DecryptOptions, w.DecryptKey)
	if err != nil {
		return inbound, fmt.Errorf("AES-GCM decryption failed: %w", err)
	}

	// 3) GOB decode => final domain T
	plaintext, err := gobDecodeToT[T](plaintextBytes)
	if err != nil {
		return inbound, fmt.Errorf("GOB decode failed: %w", err)
	}
	return plaintext, nil
}

// encryptItem is called if w.EncryptOptions.Enabled is true. Steps:
//  1. GOB encode => bytes
//  2. AES-GCM encrypt
//  3. base64 => store in T (assuming T is string).
func (w *Wire[T]) encryptItem(plaintext T) (T, error) {
	if w.EncryptOptions == nil || !w.EncryptOptions.Enabled {
		// Not encrypting => pass data
		return plaintext, nil
	}
	// 1) GOB-encode the domain object
	plaintextBytes, err := gobEncodeFromT(plaintext)
	if err != nil {
		return plaintext, fmt.Errorf("failed to GOB encode domain object: %w", err)
	}

	// 2) AES-GCM encrypt
	cipherBytes, err := encryptData(plaintextBytes, w.EncryptOptions, w.EncryptKey)
	if err != nil {
		return plaintext, fmt.Errorf("AES-GCM encryption failed: %w", err)
	}

	// 3) base64 => store as T (assuming T is string)
	base64Cipher := base64.StdEncoding.EncodeToString(cipherBytes)

	tmp := any(base64Cipher)
	finalVal, ok := tmp.(T)
	if !ok {
		return plaintext, fmt.Errorf("failed to cast base64 ciphertext to T")
	}
	return finalVal, nil
}

// gobEncodeFromT GOB-encodes the domain object T into bytes
func gobEncodeFromT[T any](val T) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(val); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// gobDecodeToT GOB-decodes from bytes => domain object T
func gobDecodeToT[T any](data []byte) (T, error) {
	var out T
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	err := dec.Decode(&out)
	return out, err
}

// encryptData does AES-GCM encryption if secOpts indicates AES-GCM. Otherwise pass plaintext
func encryptData(plaintext []byte, secOpts *relay.SecurityOptions, key string) ([]byte, error) {
	if secOpts == nil || !secOpts.Enabled || secOpts.Suite != relay.EncryptionSuite_ENCRYPTION_AES_GCM {
		return plaintext, nil
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("invalid AES key: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decryptData does AES-GCM if secOpts indicates AES-GCM. Otherwise pass ciphertext as-is
func decryptData(ciphertext []byte, secOpts *relay.SecurityOptions, key string) ([]byte, error) {
	if secOpts == nil || !secOpts.Enabled || secOpts.Suite != relay.EncryptionSuite_ENCRYPTION_AES_GCM {
		return ciphertext, nil
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("invalid AES key: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, actual := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actual, nil)
	if err != nil {
		return nil, fmt.Errorf("GCM decrypt failed: %w", err)
	}
	return plaintext, nil
}
