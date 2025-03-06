package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

func generateComplexWave(length int, frequencies []float64, amplitudes []float64, sampleRate float64) []complex128 {
	wave := make([]complex128, length)
	for i := 0; i < length; i++ {
		t := float64(i) / sampleRate
		var sum complex128
		for j, freq := range frequencies {
			sum += complex(amplitudes[j]*math.Sin(2*math.Pi*freq*t), 0)
		}
		// Add very little noise
		noise := complex(rand.NormFloat64()*0.001, rand.NormFloat64()*0.001)
		wave[i] = sum + noise
	}
	return wave
}

func calculateMSE(original, decoded []float64) float64 {
	if len(original) != len(decoded) {
		return -1 // Error: slices must have the same length
	}
	var sum float64
	for i := range original {
		diff := original[i] - decoded[i]
		sum += diff * diff
	}
	return sum / float64(len(original))
}

func findPeaks(spectrum []float64, sampleRate float64) []builder.Peak {
	var peaks []builder.Peak
	for i := 1; i < len(spectrum)-1; i++ {
		if spectrum[i] > spectrum[i-1] && spectrum[i] > spectrum[i+1] && spectrum[i] > 0 {
			freq := float64(i) * sampleRate / float64(len(spectrum)*2)
			peaks = append(peaks, builder.Peak{Freq: freq, Value: spectrum[i]})
		}
	}

	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].Value > peaks[j].Value
	})

	if len(peaks) > 5 {
		return peaks[:5]
	}
	return peaks
}

func processWaveData(item builder.WaveData) (builder.WaveData, error) {
	sampleRate := 1024.0
	compressedHex, _, waveReal, decodedFloats, _, err := builder.ProcessWaveDataWithWavelet(item.OriginalWave)
	if err != nil {
		return item, fmt.Errorf("error processing wave data: %v", err)
	}

	analysis := builder.AnalyzeWave(item.OriginalWave, sampleRate)

	item.CompressedHex = compressedHex
	item.DominantFreq = analysis["dominant_frequency"].(float64)
	item.TotalEnergy = analysis["total_energy"].(float64)
	item.CompressionRatio = 100 * (1 - float64(len(compressedHex)/2)/float64(len(item.OriginalWave)*16))
	item.MSE = calculateMSE(waveReal, decodedFloats)
	item.SNR = analysis["snr"].(float64) // Make sure this line is present

	powerSpectrum := analysis["power_spectrum"].([]float64)
	item.FrequencyPeaks = findPeaks(powerSpectrum, sampleRate)

	fmt.Printf("Processed Wave ID: %d\n", item.ID)
	fmt.Printf("  Dominant Frequency: %.2f Hz\n", item.DominantFreq)
	fmt.Printf("  Total Energy: %.2f\n", item.TotalEnergy)
	fmt.Printf("  Compression Ratio: %.2f%%\n", item.CompressionRatio)
	fmt.Printf("  Mean Squared Error: %e\n", item.MSE)
	fmt.Printf("  SNR: %.2f dB\n", item.SNR)
	fmt.Printf("  Frequency Peaks: %v\n", item.FrequencyPeaks)

	return item, nil
}

func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, item builder.WaveData) error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	id := 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sampleRate := 1024.0
			duration := 1.0
			frequencies := []float64{50, 100, 200}
			amplitudes := []float64{1.0, 0.5, 0.25}
			wave := generateComplexWave(int(sampleRate*duration), frequencies, amplitudes, sampleRate)

			item := builder.WaveData{ID: id, OriginalWave: wave}
			if err := submitFunc(ctx, item); err != nil {
				fmt.Printf("Error submitting item %d: %v\n", id, err)
				continue
			}
			id++
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := builder.NewLogger()

	plug := builder.NewPlug[builder.WaveData](
		ctx,
		builder.PlugWithAdapterFunc[builder.WaveData](plugFunc),
	)

	generator := builder.NewGenerator[builder.WaveData](
		ctx,
		builder.GeneratorWithPlug[builder.WaveData](plug),
	)

	waveEncoder := builder.NewWaveEncoder[builder.WaveData]()

	processingWire := builder.NewWire[builder.WaveData](
		ctx,
		builder.WireWithLogger[builder.WaveData](logger),
		builder.WireWithEncoder[builder.WaveData](waveEncoder),
		builder.WireWithTransformer[builder.WaveData](processWaveData),
		builder.WireWithGenerator[builder.WaveData](generator),
	)

	processingWire.Start(ctx)
	<-ctx.Done()
	processingWire.Stop()
	// Read from the Wire's output buffer
	outputBuffer := processingWire.GetOutputBuffer()
	fmt.Println("Processed Waves Summary:")

	// Decode the buffer contents
	waveDecoder := builder.NewWaveDecoder[builder.WaveData]()
	bufReader := bytes.NewReader(outputBuffer.Bytes())
	var processedWaves []builder.WaveData
	for bufReader.Len() > 0 {
		wave, err := waveDecoder.Decode(bufReader)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error decoding wave: %v\n", err)
			break
		}
		processedWaves = append(processedWaves, wave)

		// Print detailed information for each wave
		fmt.Printf("Wave ID: %d\n", wave.ID)
		fmt.Printf("  Dominant Frequency: %.2f Hz\n", wave.DominantFreq)
		fmt.Printf("  Total Energy: %.2f\n", wave.TotalEnergy)
		fmt.Printf("  Compression Ratio: %.2f%%\n", wave.CompressionRatio)
		fmt.Printf("  Mean Squared Error: %e\n", wave.MSE)
		fmt.Printf("  SNR: %.2f dB\n", wave.SNR)
		fmt.Println("  Frequency Peaks:")
		for _, peak := range wave.FrequencyPeaks {
			fmt.Printf("    Frequency: %.2f Hz, Magnitude: %.2f\n", peak.Freq, peak.Value)
		}
		fmt.Println()
	}

	// Use Filter to select only high-quality signals
	highQualityWaves := builder.Filter(processedWaves, func(wave builder.WaveData) bool {
		return wave.SNR > 3 && wave.CompressionRatio > 45 // Adjusted criteria
	})

	fmt.Printf("Total waves: %d, High-quality waves: %d\n", len(processedWaves), len(highQualityWaves))
	fmt.Println("High-quality Wave Summary:")
	for _, wave := range highQualityWaves {
		fmt.Printf("Wave ID: %d\n", wave.ID)
		fmt.Printf("  Dominant Frequency: %.2f Hz\n", wave.DominantFreq)
		fmt.Printf("  Total Energy: %.2f\n", wave.TotalEnergy)
		fmt.Printf("  Compression Ratio: %.2f%%\n", wave.CompressionRatio)
		fmt.Printf("  Mean Squared Error: %e\n", wave.MSE)
		fmt.Printf("  Signal-to-Noise Ratio: %.2f dB\n", wave.SNR)
		fmt.Println("  Frequency Peaks:")
		for _, peak := range wave.FrequencyPeaks {
			fmt.Printf("    Frequency: %.2f Hz, Magnitude: %.2f\n", peak.Freq, peak.Value)
		}
		fmt.Println()
	}

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Processing timed out.")
	} else {
		fmt.Println("Processing finished.")
	}
}
