package builder

import (
	"bytes"
	"compress/zlib"
	"encoding/gob"
	"fmt"
	"io"
	"math"
	"math/cmplx"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"github.com/mjibson/go-dsp/fft"
	"gonum.org/v1/gonum/floats"
)

type WaveData = types.WaveData
type Peak = types.Peak

// Map applies a function to each element in the slice.
func Map[T any](elems []T, f func(T) T) []T {
	return utils.Map[T](elems, f)
}

// Filter returns a new slice holding only the elements of elems that satisfy f().
func Filter[T any](elems []T, f func(T) bool) []T {
	return utils.Filter[T](elems, f)
}

// NewTransformerSequence creates a sequence of transformation functions for use in a Conduit.
func NewTransformerSequence[T any](transforms ...types.Transformer[T]) []types.Transformer[T] {
	return transforms
}

// WaveletTransformer interface for different wavelet transform implementations
type WaveletTransformer interface {
	Transform(data []float64) []float64
	InverseTransform(coeffs []float64) []float64
}

// HaarWavelet implements the Haar wavelet transform
type HaarWavelet struct{}

func (hw *HaarWavelet) Transform(data []float64) []float64 {
	n := len(data)
	coeffs := make([]float64, n)
	for i := 0; i < n; i += 2 {
		coeffs[i/2] = (data[i] + data[i+1]) / math.Sqrt2
		coeffs[n/2+i/2] = (data[i] - data[i+1]) / math.Sqrt2
	}
	return coeffs
}

func (hw *HaarWavelet) InverseTransform(coeffs []float64) []float64 {
	n := len(coeffs)
	data := make([]float64, n)
	for i := 0; i < n/2; i++ {
		data[2*i] = (coeffs[i] + coeffs[n/2+i]) / math.Sqrt2
		data[2*i+1] = (coeffs[i] - coeffs[n/2+i]) / math.Sqrt2
	}
	return data
}

func ProcessWaveDataWithWavelet(wave []complex128) (string, []complex128, []float64, []float64, string, error) {
	waveReal := make([]float64, len(wave))
	for i, v := range wave {
		waveReal[i] = real(v)
	}

	wavelet := &HaarWavelet{}
	coeffs := wavelet.Transform(waveReal)

	threshold := 0.005 * math.Max(math.Abs(floats.Max(coeffs)), math.Abs(floats.Min(coeffs)))
	for i, v := range coeffs {
		if math.Abs(v) < threshold {
			coeffs[i] = 0
		}
	}

	compressedData, err := CompressFloats(coeffs)
	if err != nil {
		return "", nil, nil, nil, "", fmt.Errorf("compression error: %v", err)
	}

	compressedHex := fmt.Sprintf("%x", compressedData)

	decompressedFloats, err := DecompressFloats(compressedData)
	if err != nil {
		return "", nil, nil, nil, "", fmt.Errorf("decompression error: %v", err)
	}

	invWaveletTransformed := wavelet.InverseTransform(decompressedFloats)

	transformedWave := make([]complex128, len(invWaveletTransformed))
	for i, v := range invWaveletTransformed {
		transformedWave[i] = complex(v, 0)
	}
	transformedEncodedWaveStr := encodeComplex(transformedWave)

	return compressedHex, wave, waveReal, decompressedFloats, transformedEncodedWaveStr, nil
}

func CompressFloats(data []float64) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	compressed := compress(buf.Bytes())
	return compressed, nil
}

func DecompressFloats(compressedData []byte) ([]float64, error) {
	decompressed := decompress(compressedData)
	buf := bytes.NewBuffer(decompressed)
	dec := gob.NewDecoder(buf)
	var data []float64
	err := dec.Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func compress(data []byte) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

func decompress(data []byte) []byte {
	var buf bytes.Buffer
	r, _ := zlib.NewReader(bytes.NewReader(data))
	io.Copy(&buf, r)
	r.Close()
	return buf.Bytes()
}

func encodeComplex(wave []complex128) string {
	return fmt.Sprintf("%v", wave)
}

func AnalyzeWave(wave []complex128, sampleRate float64) map[string]interface{} {
	analysis := make(map[string]interface{})

	fmt.Println("Analyzing wave...")
	fmt.Printf("Wave length: %d\n", len(wave))
	fmt.Printf("Sample rate: %.2f\n", sampleRate)

	// Perform FFT
	spectrum := fft.FFT(wave)
	fmt.Printf("Spectrum length: %d\n", len(spectrum))

	// Calculate power spectrum
	powerSpectrum := make([]float64, len(spectrum)/2)
	totalPower := 0.0
	maxPower := 0.0
	dominantFreqIndex := 0
	for i := range powerSpectrum {
		power := cmplx.Abs(spectrum[i]) * cmplx.Abs(spectrum[i])
		powerSpectrum[i] = power
		totalPower += power
		if power > maxPower {
			maxPower = power
			dominantFreqIndex = i
		}
	}
	analysis["power_spectrum"] = powerSpectrum

	fmt.Printf("Total power: %.2f\n", totalPower)
	fmt.Printf("Max power: %.2f\n", maxPower)
	fmt.Printf("Dominant frequency index: %d\n", dominantFreqIndex)

	// Find dominant frequency
	dominantFreq := float64(dominantFreqIndex) * sampleRate / float64(len(wave))
	analysis["dominant_frequency"] = dominantFreq
	fmt.Printf("Dominant frequency: %.2f Hz\n", dominantFreq)

	// Calculate total energy (in time domain)
	totalEnergy := 0.0
	for _, v := range wave {
		totalEnergy += cmplx.Abs(v) * cmplx.Abs(v)
	}
	analysis["total_energy"] = totalEnergy
	fmt.Printf("Total energy: %.2f\n", totalEnergy)

	// Calculate signal power (assuming the dominant frequency is the signal)
	signalPower := maxPower
	noisePower := totalPower - signalPower

	fmt.Printf("Signal power: %.2f\n", signalPower)
	fmt.Printf("Noise power: %.2f\n", noisePower)

	snr := 10 * math.Log10(signalPower/noisePower)
	analysis["snr"] = snr

	return analysis
}
